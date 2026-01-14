<script lang="ts">
  import { RARITY_MAP, type RarityValue } from '$lib/rarity';

  export let value: string = '';
  export let prefix = 'Rarity';
  export let showPrefix = true;
  export let boldPrefix = true;
  export let asBadge = false;
  export let className = '';

  // Keep these reactive so they update when the `value` prop changes
  let key = (value ?? '').toLowerCase() as RarityValue;
  $: key = (value ?? '').toLowerCase() as RarityValue;

  let rarity = RARITY_MAP[key] ?? { label: value ?? '', class: 'text-gray-400' };
  $: rarity = RARITY_MAP[key] ?? { label: value ?? '', class: 'text-gray-400' };

  const badgeBase =
    'inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-sm font-medium ring-1 ring-black/5 ' +
    'bg-gray-100 dark:bg-white/5';
</script>

{#if asBadge}
  <span class={`${badgeBase} ${className}`}>
    {#if showPrefix}<span class={boldPrefix ? 'font-semibold opacity-80' : 'opacity-80'}>{prefix}:</span>{/if}
    <span class={rarity.class}>{rarity.label}</span>
  </span>
{:else}
  <span class={`inline-flex items-baseline gap-1 ${className}`}>
    {#if showPrefix}
      <span class={boldPrefix ? 'font-semibold' : ''}>{prefix}:</span>
    {/if}
    <span class={rarity.class}>{rarity.label}</span>
  </span>
{/if}
