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

		// Build filter parts separately for proper parameterization
		let filterParts = [`endTime > {:endTime}`];
		let filterParams: Record<string, any> = { endTime: new Date(date).toISOString() };

		if (searchQuery != '') {
			filterParts.push(`itemName ~ {:searchQuery}`);
			filterParams.searchQuery = searchQuery + '%';
		}

		// Do not sort by favourites; always sort by end time. If "showFavouritesOnly" is enabled,
		// filter to auctions that have a favourite record for the current user.
		if (showFavouritesOnly && $user) {
			// PocketBase supports filtering on expanded relations when using the relation name.
			filterParts.push(`favourites.id ?= {:userId}`);
			filterParams.userId = $user.id;
		}

		const filterString = pb.filter(filterParts.join(' && '), filterParams);
		console.debug('Fetching items with filter:', filterString);

		const records = await pb
			.collection('auctions')
			.getList(page, itemsPerPage, {
				sort: '-endTime,id',
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

	function handleFavouriteToggled() {
		// Refetch items to update the list when a favourite is toggled
		fetchItems(currentPage).then(subscribeToCurrentPage);
	}

	fetchItems(currentPage).then(subscribeToCurrentPage);
</script>

{#if $user}
	<div class="space-y-6">
		<!-- Header Section -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h1 class="text-2xl font-bold mb-4 flex items-center gap-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
				</svg>
				Active Auctions
			</h1>
			<div class="flex flex-col sm:flex-row gap-4 items-start sm:items-center justify-between">
				<div class="form-control w-full max-w-md">
					<div class="input-group">
						<input
							type="text"
							placeholder="Search auctions by name..."
							class="input input-bordered w-full shadow-sm"
							bind:value={searchQuery}
							on:input={handleSearch}
						/>
						<button class="btn btn-square btn-primary" aria-label="Search">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
							</svg>
						</button>
					</div>
				</div>
				<label class="cursor-pointer flex items-center gap-3 bg-base-100 px-4 py-3 rounded-box shadow-sm border border-base-content/10 hover:border-primary transition-colors">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-error" viewBox="0 0 24 24" fill="currentColor">
						<path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 6 4 4 6.5 4c1.74 0 3.41.81 4.5 2.09C12.09 4.81 13.76 4 15.5 4 18 4 20 6 20 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
					</svg>
					<span class="label-text font-medium">Favourites only</span>
					<input
						type="checkbox"
						class="toggle toggle-error"
						bind:checked={showFavouritesOnly}
						on:change={handleToggleFavourites}
					/>
				</label>
			</div>
		</div>

		<!-- Auctions Grid -->
		{#if items.items.length > 0}
			<div class="grid grid-cols-1 gap-6 md:grid-cols-2 lg:grid-cols-3">
				{#each items.items as item}
					<AuctionItem {item} on:favouriteToggled={handleFavouriteToggled} />
				{/each}
			</div>
		{:else}
			<div class="flex flex-col items-center justify-center py-16 bg-base-200 rounded-box border-2 border-dashed border-base-content/20">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 opacity-30 mb-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4" />
				</svg>
				<p class="text-xl opacity-70">No auctions found</p>
				<p class="text-sm opacity-50 mt-2">Try adjusting your search or filters</p>
			</div>
		{/if}

		<!-- Pagination -->
		{#if items.totalPages > 0}
			<div class="flex justify-center">
				<div class="join shadow-lg">
					<button
						class="join-item btn btn-lg"
						tabindex="-1"
						on:click={() => changePage(currentPage - 1)}
						disabled={currentPage === 1}
						aria-label="Previous page"
					>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
					</button>
					<button class="join-item btn btn-lg pointer-events-none">
						Page <span class="font-bold mx-1">{currentPage}</span> of <span class="font-bold mx-1">{items.totalPages}</span>
					</button>
					<button
						class="join-item btn btn-lg"
						tabindex="-1"
						on:click={() => changePage(currentPage + 1)}
						disabled={currentPage === items.totalPages || items.totalPages === 0}
						aria-label="Next page"
					>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</button>
				</div>
			</div>
		{/if}
	</div>
{:else}
	<div class="hero min-h-[50vh] bg-base-200 rounded-box shadow-xl">
		<div class="hero-content text-center">
			<div class="max-w-md">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 mx-auto mb-4 opacity-50" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
				</svg>
				<h1 class="text-3xl font-bold mb-3">Welcome to Auction House</h1>
				<p class="text-base opacity-80 mb-4">Please log in to view and participate in auctions</p>
			</div>
		</div>
	</div>
{/if}
