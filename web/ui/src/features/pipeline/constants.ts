import type { ReleaseTag } from '@/types';

export const RELEASE_TAGS: { value: ReleaseTag; label: string }[] = [
  { value: 'stable', label: 'Stable' },
  { value: 'hot-fix', label: 'Hot Fix' },
  { value: 'night-build', label: 'Night Build' },
  { value: 'develop', label: 'Develop' },
];
