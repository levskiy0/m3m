import type { GoalType } from '@/types';

export const GOAL_TYPES: { value: GoalType; label: string }[] = [
  { value: 'counter', label: 'Counter' },
  { value: 'daily_counter', label: 'Daily Counter' },
];
