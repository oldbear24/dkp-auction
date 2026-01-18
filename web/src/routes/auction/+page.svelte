<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { page } from '$app/state';
  import pb, { subscribeToAuctionUpdate } from '$lib/pocketbase';
  import AuctionItem from '../../components/AuctionItem.svelte';
  import type { RecordModel } from 'pocketbase';

  let item: RecordModel | null = null;
  let loading = true;

  $: id = page.url.searchParams.get('id') ?? '';

  onMount(async () => {
    console.debug('Loading auction with id:', id);
    try {
      item = await pb.collection('auctions').getOne(id);
      await subscribeToAuctionUpdate(id, (updatedRecord) => {
        console.debug('Received auction update:', updatedRecord);
        item = updatedRecord;
      });
    } catch (err) {
      console.error('Failed to load auction', err);
      item = null;
    } finally {
      loading = false;
    }
  });

onDestroy(() => {
    if (id) {
      pb.collection('auctions').unsubscribe(id);
    }
  });
</script>
  {#if loading}
    <div class="flex items-center justify-center h-64">Loading auction...</div>
  {:else}
    {#if item}
      <div class="container mx-auto px-4 py-6">
        <div class="mb-4">
          <a href="/" class="link">← Back to auctions</a>
        </div>
        <AuctionItem {item} />
      </div>
    {:else}
      <div class="flex items-center justify-center h-64">Auction not found</div>
    {/if}
  {/if}
