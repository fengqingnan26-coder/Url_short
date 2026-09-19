"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { useAuth } from "@/components/context";

export default function AuthLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  const router = useRouter();
  const { isAuth } = useAuth();

  useEffect(() => {
    if (isAuth) {
      router.push("/");
    }
  }, [isAuth, router]);

  if (isAuth) {
    return null;
  }

  return <>{children}</>;
}
