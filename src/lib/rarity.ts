export type RarityValue =
  | 'common'
  | 'uncommon'
  | 'rare'
  | 'rare_t2'
  | 'epic'
  | 'epic_t2'
  | 'epic_t3'
  | 'heroic'
  | 'artifact';

export const RARITY_MAP: Record<RarityValue, { label: string; class: string }> = {
  common:    { label: 'Common',    class: 'text-gray-500 dark:text-gray-400' },
  uncommon:  { label: 'Uncommon',  class: 'text-emerald-500 dark:text-emerald-400' },
  rare:      { label: 'Rare',      class: 'text-blue-500 dark:text-blue-400' },
  rare_t2:   { label: 'Rare T2',   class: 'text-indigo-500 dark:text-indigo-400' },
  epic:      { label: 'Epic',      class: 'text-violet-500 dark:text-violet-400' },
  epic_t2:   { label: 'Epic T2',   class: 'text-purple-500 dark:text-purple-400' },
  epic_t3:   { label: 'Epic T3',   class: 'text-fuchsia-500 dark:text-fuchsia-400' },
  heroic:    { label: 'Heroic',    class: 'text-amber-500 dark:text-amber-400' },
  artifact:  { label: 'Artifact',  class: 'text-rose-500 dark:text-rose-400' }
};

export const RARITY_OPTIONS: RarityValue[] = Object.keys(RARITY_MAP) as RarityValue[];
