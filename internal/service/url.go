package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jekyulll/url_shortener/config"
	"github.com/jekyulll/url_shortener/internal/cache"
	"github.com/jekyulll/url_shortener/internal/dto"
	"github.com/jekyulll/url_shortener/internal/model"
	"github.com/jekyulll/url_shortener/internal/repository"
	"github.com/jekyulll/url_shortener/pkg/filter"
	"github.com/jekyulll/url_shortener/pkg/shortcode"
	"gorm.io/gorm"
)

type CodeStatus int

const (
	CodeAvailable CodeStatus = iota // 全新可用
	CodeExpired                     // 存在但已过期
	CodeInUse                       // 存在且有效
)

type URLCacher interface {
	SetURL(ctx context.Context, url model.URL) error
	GetURL(ctx context.Context, shortCode string) (*model.URL, error) // TODO 返回 string
	DelURL(ctx context.Context, shortCode string) error
	IncreViews(ctx context.Context, shortCode string) error
	ScanViews(ctx context.Context, cursor uint64, batchSize int64) (keys []string, nextCursor uint64, err error)
	GetViews(ctx context.Context, shortCode string) (int, error)
	DelViews(ctx context.Context, shortCode string) error
}

type ShortCodeGenerator interface {
	GenerateShortCode() (string, error)
}

type URLService struct {
	repo               repository.URLRepository
	filter             filter.BloomFilter
	shortCodeGenerator ShortCodeGenerator
	defaultDuration    time.Duration
	cache              URLCacher
	baseURL            string
}

func NewURLService(repo repository.URLRepository, filter filter.BloomFilter, generator ShortCodeGenerator, cache URLCacher, cfg config.AppConfig) *URLService {
	// 启动时加载所有有效短码到过滤器
	if urls, err := repo.GetAllActiveURLs(context.Background()); err == nil {
		for _, url := range urls {
			filter.Add(url.ShortCode)
		}
	}
	return &URLService{
		repo:               repo,
		filter:             filter,
		shortCodeGenerator: generator,
		cache:              cache,
		defaultDuration:    cfg.DefaultDuration,
		baseURL:            cfg.BaseURL,
	}
}

// GetURLs implements api.URLServicer.
func (s *URLService) GetURLs(ctx context.Context, req dto.GetURLsRequest) (*dto.GetURLsResponse, error) {
	rows, err := s.repo.GetURLsByUserID(ctx, int32(req.UserID), int32(req.Size), int32(req.Page-1))
	if err != nil {
		return nil, err
	}
	items := make([]dto.FullURL, len(rows))
	total := 0

	for i := range rows {
		// TODO 检查 total
		row := rows[i]
		views, err := s.cache.GetViews(ctx, row.ShortCode)
		if err != nil {
			return nil, err
		}
		row.Views += int32(views)
		items[i] = dto.FullURL{
			ID:          int(row.ID),
			OriginalURL: row.OriginalURL,
			ShortURL:    fmt.Sprintf("%s/%s", s.baseURL, row.ShortCode),
			ExpiredAt:   row.ExpiredAt,
			IsCustom:    row.IsCustom,
			Views:       uint(row.Views),
		}
		total = int(row.Views)
	}
	resp := dto.GetURLsResponse{
		Items: items,
		Total: total,
	}
	return &resp, nil
}

// DefaultURL implements api.URLServicer.
func (s *URLService) DefaultURL(ctx context.Context) error {
	// panic("unimplemented")
	// TODO 函数签名需要修改，默认导航到一个404页面
	return nil
}

// DeleteURL implements api.URLServicer.
func (s *URLService) DeleteURL(ctx context.Context, shortCode string) error {
	if err := s.repo.DeleteURLByShortCode(ctx, shortCode); err != nil {
		return err
	}
	if err := s.cache.DelURL(ctx, shortCode); err != nil {
		return err
	}
	if err := s.cache.DelViews(ctx, shortCode); err != nil {
		return err
	}
	return nil
}

// IncreViews implements api.URLServicer.
func (s *URLService) IncreViews(ctx context.Context, shortCode string) error {
	return s.cache.IncreViews(ctx, shortCode)
}

// UpdateURLDuration implements api.URLServicer.
func (s *URLService) UpdateURLDuration(ctx context.Context, req dto.UpdateURLDurationReq) error {
	// FIXME 是否会导致缓存里的持续时间还在？
	// 因此此处删除缓存，旁路缓存模式
	if err := s.cache.DelURL(ctx, req.Code); err != nil {
		return fmt.Errorf("failed to delete cache: %v", err.Error())
	}
	return s.repo.UpdateURLExpiredByShortCode(ctx, req.Code, req.ExpiredAt)
}

// 如出错返回 err，如短链接已存在，返回预定义错误 ErrShortCodeTaken
func (s *URLService) CreateURL(ctx context.Context, req dto.CreateURLRequest) (*dto.CreateURLResponse, error) {
	var expiredAt time.Time
	// 1. 决定要用的短码：优先用用户自己的，其次自动生成
	code := req.CustomeCode
	var err error
	if code == "" {
		code, err = s.getShortCode(ctx, 0)
		if err != nil {
			return nil, err
		}
	} else {
		// 用户定制，先验证它是否可用
		status, err := s.CheckShortCode(ctx, code)
		if err != nil {
			return nil, fmt.Errorf("check custom shortcode: %w", err)
		}
		if status == CodeInUse {
			return nil, ErrShortCodeTaken
		}
	}
	// 2. 写入布隆过滤器
	s.filter.Add(code)
	// 3. 组装要入库的 URL 实体
	if req.Duration == nil {
		expiredAt = time.Now().Add(s.defaultDuration)
	} else {
		expiredAt = time.Now().Add(time.Hour * time.Duration(*req.Duration))
	}
	u := &model.URL{
		OriginalURL: req.OriginalURL,
		ShortCode:   code,
		IsCustom:    req.CustomeCode != "",
		UserID:      uint64(req.UserID),
		ExpiredAt:   expiredAt,
	}
	// 4. 处理过期时间：不传的话使用默认有效期
	if req.Duration == nil {
		u.ExpiredAt = time.Now().Add(s.defaultDuration)
	} else {
		u.ExpiredAt = time.Now().Add(time.Duration(*req.Duration) * time.Hour)
	}
	// 5. 写库
	if err := s.repo.UpsertURL(ctx, u); err != nil {
		return nil, fmt.Errorf("create url record: %w", err)
	}
	// 6. 异步写缓存
	go func() {
		if err := s.cache.SetURL(context.Background(), *u); err != nil {
			log.Printf("failed to set cache: %v", err)
		}
	}()
	// 6. 返回给上层
	return &dto.CreateURLResponse{
		ShortUrl: s.baseURL + "/" + u.ShortCode,
		//ExpiredAt: u.ExpiredAt,
	}, nil
}

func (s *URLService) GetURL(ctx context.Context, shortCode string) (string, error) {
	// 1. 查找布隆过滤器
	if !s.filter.Exists(shortCode) { // 布隆过滤器中不存在，则一定不存在。
		return "", nil
	}
	// 2. 访问缓存
	url, err := s.cache.GetURL(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if url != nil {
		return url.OriginalURL, nil
	}
	// 3. 缓存中不存在，访问数据库
	url, err = s.repo.GetURLByShortCode(ctx, shortCode)
	if err != nil {
		return "", err
	}
	if url == nil { // 不是查询出错，而是数据库没该数据
		return "", nil
	}
	// 4. 存入缓存
	go func() {
		if err := s.cache.SetURL(context.Background(), *url); err != nil {
			log.Printf("failed to set cache: %v", err)
		}
	}()
	return url.OriginalURL, nil
}

// @pragma n:重试次数
func (s *URLService) getShortCode(ctx context.Context, n int) (string, error) {
	if n > 5 {
		return "", errors.New("retry too many times")
	}
	code, err := s.shortCodeGenerator.GenerateShortCode()
	if err != nil {
		return "", err
	}
	status, err := s.CheckShortCode(ctx, code)
	if err != nil {
		return "", err
	}
	if status == CodeAvailable || status == CodeExpired {
		return code, nil
	}
	// 递归调用
	return s.getShortCode(ctx, n+1)
}

func (s *URLService) DeleteAllExpired(ctx context.Context) error {
	return s.repo.DeleteAllExpired(ctx)
}

// TODO 短链接会过期，需要定时重建布隆过滤器（扫描整个数据库，很麻烦）

// CheckShortCode NEW: 从 repo 层提到 service, 与布隆过滤器的判断整合
// TODO 增加缓存的逻辑
// 如果存在于数据库、但是过期了，也视为合法 —— 生成新链接，写入数据库覆盖之前的
func (s *URLService) CheckShortCode(ctx context.Context, shortCode string) (CodeStatus, error) {
	if !s.filter.Exists(shortCode) { // 布隆过滤器中不存在，则一定不存在。
		return CodeAvailable, nil
	}
	// 布隆过滤器中存在，仍然可能不存在。判断是否在数据库中
	url, err := s.repo.GetURLByShortCode(ctx, shortCode)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return CodeAvailable, nil // 不存在
	}
	if err != nil {
		return CodeInUse, err
	}
	if url.IsExpired() {
		return CodeExpired, nil
	}
	return CodeInUse, nil
}

func (s *URLService) SyncViewsToDB(ctx context.Context) error {
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = s.cache.ScanViews(ctx, cursor, 100)
		if err != nil {
			return err
		}

		for _, key := range keys {
			views, err := s.cache.GetViews(ctx, key)
			if err != nil {
				return err
			}

			if views == 0 {
				continue
			}

			if err := s.cache.DelViews(ctx, key); err != nil {
				return err
			}

			shortCode := strings.Split(key, ":")[1]

			if err := s.repo.UpdateViewsByShortCode(ctx, shortCode, int32(views)); err != nil {
				return err
			}
		}

		if cursor == 0 {
			break
		}
	}

	return nil
}

var _ URLCacher = (*cache.RedisCache)(nil)
var _ ShortCodeGenerator = (*shortcode.RandomShortCodeGeneratorImpl)(nil)
var _ ShortCodeGenerator = (*shortcode.SnowflakeShortCodeGenerator)(nil)
