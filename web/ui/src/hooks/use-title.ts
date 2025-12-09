import { useEffect } from 'react';

const APP_NAME = 'M3M';

export function useTitle(title: string | undefined | null) {
  useEffect(() => {
    const prevTitle = document.title;
    document.title = title ? `${title} | ${APP_NAME}` : APP_NAME;
    return () => {
      document.title = prevTitle;
    };
  }, [title]);
}
