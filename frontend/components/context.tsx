"use client";

import { createContext, use, useEffect, useState } from "react";
import { isTokenExpired } from "./token";

type AuthProviderProps = {
  children: React.ReactNode;
};

type AuthProviderState = {
  token: string;
  email: string;
  isAuth: boolean;
  userID: number;
  mounted: boolean;
  setAuth: (token: string, email: string, userID: number) => void;
};

const AuthProviderContext = createContext<AuthProviderState>({
  token: "",
  email: "",
  userID: 0,
  isAuth: false,
  mounted: false,
  setAuth: () => null,
});

// 统一从 localStorage 读取，SSR 和 CSR 首次渲染都返回空值避免 hydration mismatch
function readStorage(key: string): string {
  try {
    if (typeof window === "undefined") return "";
    return window.localStorage.getItem(key) ?? "";
  } catch {
    return "";
  }
}

function writeStorage(key: string, value: string) {
  try {
    if (typeof window !== "undefined") {
      window.localStorage.setItem(key, value);
    }
  } catch {
    /* ignore */
  }
}

export function AuthProvider({ children }: AuthProviderProps) {
  // mounted 前 SSR 和 CSR 都用空值渲染，保持 hydration 一致
  const [mounted, setMounted] = useState(false);
  const [email, setEmail] = useState("");
  const [token, setToken] = useState("");
  const [userID, setUserID] = useState("");

  useEffect(() => {
    // 客户端挂载后才从 localStorage 读取真实值
    setEmail(readStorage("email"));
    setToken(readStorage("token"));
    setUserID(readStorage("user_id"));
    setMounted(true);
  }, []);

  const isAuth = mounted && !isTokenExpired(token) && parseInt(userID) !== 0 && email !== "";

  const value = {
    token,
    email,
    userID: parseInt(userID),
    isAuth,
    mounted,
    setAuth: (newToken: string, newEmail: string, newUserID: number) => {
      setToken(newToken);
      setEmail(newEmail);
      setUserID(String(newUserID));
      writeStorage("token", newToken);
      writeStorage("email", newEmail);
      writeStorage("user_id", String(newUserID));
    },
  };

  return <AuthProviderContext value={value}>{children}</AuthProviderContext>;
}

export const useAuth = () => {
  const context = use(AuthProviderContext);

  if (context === undefined) {
    throw new Error("useAuth must be use within a AuthProvider");
  }

  return context;
};
