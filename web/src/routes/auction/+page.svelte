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
    <div class="flex flex-col items-center justify-center py-16 bg-base-200 rounded-box shadow-lg">
			<span class="loading loading-spinner loading-lg text-primary"></span>
			<p class="text-lg opacity-70 mt-4">Loading auction details...</p>
		</div>
  {:else}
    {#if item}
      <div class="space-y-6">
				<!-- Back Button -->
        <div class="bg-base-200 rounded-box p-4 shadow-lg border border-base-content/10">
          <a href="/" class="btn btn-ghost gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
						Back to Auctions
					</a>
        </div>
				<!-- Auction Item -->
        <AuctionItem {item} />
      </div>
    {:else}
      <div class="hero min-h-[60vh] bg-base-200 rounded-box shadow-xl">
				<div class="hero-content text-center">
					<div class="max-w-md">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-24 w-24 mx-auto mb-6 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<h1 class="text-3xl font-bold mb-4">Auction Not Found</h1>
						<p class="text-lg opacity-80 mb-6">The auction you're looking for doesn't exist or has been removed.</p>
						<a href="/" class="btn btn-primary gap-2">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
							</svg>
							Back to Auctions
						</a>
					</div>
				</div>
			</div>
    {/if}
  {/if}
