import { useState, useCallback } from 'react';
import { slugify } from '@/lib/utils';

interface UseAutoSlugOptions {
  separator?: string;
}

interface UseAutoSlugReturn {
  name: string;
  slug: string;
  setName: (value: string) => void;
  setSlug: (value: string) => void;
  reset: () => void;
}

/**
 * Hook for auto-generating slug from name with manual override support.
 * Once user manually edits the slug, auto-generation stops.
 */
export function useAutoSlug(options: UseAutoSlugOptions = {}): UseAutoSlugReturn {
  const { separator = '-' } = options;

  const [name, setNameState] = useState('');
  const [slug, setSlugState] = useState('');
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);

  const setName = useCallback((value: string) => {
    setNameState(value);
    if (!slugManuallyEdited) {
      setSlugState(slugify(value, separator));
    }
  }, [slugManuallyEdited, separator]);

  const setSlug = useCallback((value: string) => {
    setSlugState(slugify(value, separator));
    setSlugManuallyEdited(true);
  }, [separator]);

  const reset = useCallback(() => {
    setNameState('');
    setSlugState('');
    setSlugManuallyEdited(false);
  }, []);

  return {
    name,
    slug,
    setName,
    setSlug,
    reset,
  };
}
