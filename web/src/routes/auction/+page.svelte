<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { page } from '$app/state';
	import pb, { subscribeToAuctionUpdate } from '$lib/pocketbase';
	import AuctionItem from '../../components/AuctionItem.svelte';
	import type { RecordModel } from 'pocketbase';

	let item: RecordModel | null = null;
	let loading = true;
	let mounted = false;
	let currentSubscriptionId: string | null = null;

	$: id = page.url.searchParams.get('id') ?? '';

	// Reactive statement to handle id changes
	$: if (mounted && id) {
		loadAuction(id);
	}

	async function loadAuction(auctionId: string) {
		loading = true;
		console.debug('Loading auction with id:', auctionId);

		// Unsubscribe from previous auction if exists
		if (currentSubscriptionId && currentSubscriptionId !== auctionId) {
			pb.collection('auctions').unsubscribe(currentSubscriptionId);
		}

		try {
			item = await pb.collection('auctions').getOne(auctionId);
			await subscribeToAuctionUpdate(auctionId, (updatedRecord) => {
				console.debug('Received auction update:', updatedRecord);
				item = updatedRecord;
			});
			currentSubscriptionId = auctionId;
		} catch (err) {
			console.error('Failed to load auction', err);
			item = null;
		} finally {
			loading = false;
		}
	}

	onMount(async () => {
		mounted = true;
		if (id) {
			await loadAuction(id);
		}
	});

	onDestroy(() => {
		if (currentSubscriptionId) {
			pb.collection('auctions').unsubscribe(currentSubscriptionId);
		}
	});
</script>

{#if loading}
	<div class="flex h-64 items-center justify-center">Loading auction...</div>
{:else if item}
	<div class="container mx-auto px-4 py-6">
		<div class="mb-4">
			<a href="/" class="link">← Back to auctions</a>
		</div>
		<AuctionItem {item} />
	</div>
{:else}
	<div class="flex h-64 items-center justify-center">Auction not found</div>
{/if}
