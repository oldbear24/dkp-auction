<script lang="ts">
	import type { ListResult, RecordModel } from 'pocketbase';
	import pb, { subscribeToAuctionUpdate, unsubscribeFromAuctionUpdates } from '../lib/pocketbase';
	import AuctionItem from '../components/AuctionItem.svelte';
	import { user } from '$lib/stores/store';
	const itemsPerPage = 9;
	let items: ListResult<RecordModel> = {
		page: 1,
		perPage: itemsPerPage,
		totalItems: 0,
		totalPages: 0,
		items: []
	};
	let currentPage = 1;
	let searchQuery = '';
	let showFavouritesOnly = false;
	async function fetchItems(page: number) {
		const date = new Date();
		date.setDate(date.getDate() - 2);

		// Build filter parts array
		const filterParts: string[] = [];
		const filterParams: Record<string, any> = {};

		// Add date filter
		filterParams.endTime = new Date(date).toISOString();
		filterParts.push('endTime > {:endTime}');

		// Add search query filter if present
		if (searchQuery !== '') {
			filterParams.searchQuery = searchQuery + '%';
			filterParts.push('itemName ~ {:searchQuery}');
		}

		// Add favourites filter if enabled
		if (showFavouritesOnly && $user) {
			filterParams.userId = $user.id;
			filterParts.push('favourites_via_auction.user = {:userId}');
		}

		// Combine filter parts with AND operator
		const filterExpression = filterParts.join(' && ');
		const filterString = pb.filter(filterExpression, filterParams);

		console.debug('Fetching items with filter:', filterString);
		const records = await pb
			.collection('auctions')
			.getList(page, itemsPerPage, {
				expand: 'favourites_via_auction',
				sort: '-endTime',
				filter: filterString
			});
		items = records;
	}

	function updateItem(record: RecordModel) {
		const index = items.items.findIndex((item) => item.id === record.id);
		console.debug('Updating item:', record.id, index);

		if (index !== -1) {
			items.items[index] = record;
			items.items = [...items.items]; // Trigger reactivity
		}
	}

	async function subscribeToCurrentPage() {
		await unsubscribeFromAuctionUpdates();
		console.debug('Subscribing to updates for current page');
		const recordIds = items.items.map((item) => item.id);
		recordIds.forEach((recordId) => subscribeToAuctionUpdate(recordId, updateItem));
	}

	function changePage(page: number) {
		currentPage = page;
		fetchItems(currentPage).then(subscribeToCurrentPage);
	}

	function handleSearch() {
		fetchItems(1).then(subscribeToCurrentPage);
	}

	function handleToggleFavourites() {
		// reset to first page when toggling
		currentPage = 1;
		fetchItems(1).then(subscribeToCurrentPage);
	}

	fetchItems(currentPage).then(subscribeToCurrentPage);
</script>

{#if $user}
	<div class="container mx-auto overflow-x-auto px-4 pb-4">
		<div class="mb-4 flex items-center justify-between">
			<input
				type="text"
				placeholder="Search auctions..."
				class="input input-bordered w-full max-w-xs"
				bind:value={searchQuery}
				on:input={handleSearch}
			/>
			<label class="ml-4 flex items-center space-x-2">
				<input
					type="checkbox"
					class="toggle"
					bind:checked={showFavouritesOnly}
					on:change={handleToggleFavourites}
				/>
				<span>Show favourites only</span>
			</label>
		</div>
		<div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
			{#each items.items as item}
				<AuctionItem {item} />
			{/each}
		</div>
	</div>
	<div class="join flex justify-center pt-2">
		<button
			class="join-item btn"
			tabindex="-1"
			on:click={() => changePage(currentPage - 1)}
			disabled={currentPage === 1}>«</button
		>
		<button class="join-item btn" tabindex="-1">Page {currentPage} of {items.totalPages}</button>
		<button
			class="join-item btn"
			tabindex="-1"
			on:click={() => changePage(currentPage + 1)}
			disabled={currentPage === items.totalPages || items.totalPages === 0}>»</button
		>
	</div>
{:else}
	<div class="flex h-screen items-center justify-center">
		<p class="text-xl">You are not logged in. Please log in to view the auction items.</p>
	</div>
{/if}
