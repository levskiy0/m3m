import type { ModelData, FilterCondition } from '@/types';

// Tab types for file-manager style interface
export type TabType = 'table' | 'view' | 'edit' | 'create';

export interface Tab {
  id: string;
  type: TabType;
  title: string;
  data?: ModelData;
  formData?: Record<string, unknown>;
  fieldErrors?: Record<string, string>;
  hasChanges?: boolean;
}

export interface ActiveFilter {
  field: string;
  operator: FilterCondition['operator'];
  value: unknown;
}
