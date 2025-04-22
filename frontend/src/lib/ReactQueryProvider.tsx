"use client"; // クライアントコンポーネントとしてマーク

import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ReactQueryDevtools } from '@tanstack/react-query-devtools';

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: {
        // デフォルトのクエリオプション (必要に応じて設定)
        staleTime: 60 * 1000, // 1分間は stale とみなさない
      },
    },
  });
}

let browserQueryClient: QueryClient | undefined = undefined;

function getQueryClient() {
  if (typeof window === 'undefined') {
    // Server: 新しいクライアントを常に作成
    return makeQueryClient();
  } else {
    // Browser: クライアントが存在しなければ作成
    if (!browserQueryClient) browserQueryClient = makeQueryClient();
    return browserQueryClient;
  }
}

type Props = {
  children: React.ReactNode;
};

const ReactQueryProvider: React.FC<Props> = ({ children }) => {
  // Server Components でレンダリングされた際に React Query の重複を避ける
  const queryClient = getQueryClient();

  return (
    <QueryClientProvider client={queryClient}>
      {children}
      <ReactQueryDevtools initialIsOpen={false} /> {/* 開発モードでのみ表示される */} 
    </QueryClientProvider>
  );
};

export default ReactQueryProvider; 