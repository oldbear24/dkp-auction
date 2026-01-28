<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import pb from '$lib/pocketbase';
	import type { RecordModel, RecordSubscription } from 'pocketbase';
	import AuthGuard from '../../components/AuthGuard.svelte';

	let auctionResults = writable<RecordModel[]>([]);
	const auctionFields = 'id,expand.auction.itemName,expand.auction.description,resolved,expand.resolvedBy.name,created,expand.auction.expand.winner.name';
	const auctionExpand = 'resolvedBy,auction,auction.winner';
	let showUnresolvedOnly = writable(true);

	async function fetchAuctionResults() {
		try {
			let filterQuery = '';
			if ($showUnresolvedOnly) {
				filterQuery = 'resolved = false';
			}
			const response = await pb.collection('auctionsResult').getFullList({
				sort: '-created',
				filter: filterQuery,
				fields:  auctionFields,
				expand: auctionExpand
			});
			console.debug('Filter:', filterQuery);
			auctionResults.set(response);
			console.debug(response);
		} catch (error) {
			console.error('Error fetching auction results:', error);
		}
	}

	onMount(() => {
		fetchAuctionResults();
		pb.collection('auctionsResult').subscribe('*', updateTableFromRT);
	});
	onDestroy(() => {
		pb.collection('auctionsResult').unsubscribe('*');
	});

	async function updateTableFromRT(recordSub: RecordSubscription<RecordModel>) {
        console.debug(recordSub)
		switch (recordSub.action) {
			case 'create':
				pb.collection('auctionsResult')
					.getOne(recordSub.record.id, { fields:auctionFields ,expand: auctionExpand })
					.then((record) => {
						auctionResults.set([record, ...$auctionResults]);
					});
				break;
			case 'update':
				pb.collection('auctionsResult')
					.getOne(recordSub.record.id, {fields:auctionFields , expand: auctionExpand })
					.then((record) => {
						auctionResults.update((results) =>
							results.map((result) =>
								result.id === record.id ? record : result
							)
						);
					});
                    break;
			case 'delete':
				auctionResults.set($auctionResults.filter((result) => result.id !== recordSub.record.id));
			default:
		}
	}
	function resolveAuction(id:string,isResolved:true) {
		if(isResolved)return
		try
		{
			pb.send("/api/resolve-auction/"+id,{
				method:"POST"
			})
		}
		catch(err){
			console.error(err)
		}
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
		<h1 class="text-3xl font-bold mb-2 flex items-center gap-3">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			Auction Results
		</h1>
		<p class="text-sm opacity-70">Review and manage completed auction outcomes</p>
	</div>

	<!-- Filter Options -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
		<h2 class="text-lg font-semibold mb-3 flex items-center gap-2">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 4a1 1 0 011-1h16a1 1 0 011 1v2.586a1 1 0 01-.293.707l-6.414 6.414a1 1 0 00-.293.707V17l-4 4v-6.586a1 1 0 00-.293-.707L3.293 7.293A1 1 0 013 6.586V4z" />
			</svg>
			Filter Options
		</h2>
		<label class="cursor-pointer flex items-center gap-3 bg-base-100 px-4 py-3 rounded-box shadow-sm border border-base-content/10 hover:border-primary transition-colors w-fit">
			<input
				type="checkbox"
				bind:checked={$showUnresolvedOnly}
				on:change={fetchAuctionResults}
				class="toggle toggle-info toggle-lg"
			/>
			<div class="flex items-center gap-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span class="font-medium">Show Unresolved Only</span>
			</div>
		</label>
	</div>

	<!-- Results Table -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
		<div class="overflow-x-auto rounded-box border border-base-content/10 bg-base-100 shadow-inner">
			<table class="table table-zebra">
				<thead class="bg-base-300">
					<tr>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 7l-8-4-8 4m16 0l-8 4m8-4v10l-8 4m0-10L4 7m8 4v10M4 7v10l8 4" />
								</svg>
								Auction Item
							</div>
						</th>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
								</svg>
								Description
							</div>
						</th>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
								</svg>
								Status
							</div>
						</th>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
								</svg>
								Resolved By
							</div>
						</th>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />
								</svg>
								Winner
							</div>
						</th>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
								Created
							</div>
						</th>
					</tr>
				</thead>
				<tbody>
					{#if $auctionResults.length > 0}
						{#each $auctionResults as result}
							<tr class="hover">
								<td class="font-semibold">{result.expand?.auction.itemName}</td>
								<td class="max-w-md truncate text-sm opacity-80">{result.expand?.auction.description}</td>
								<td>
									<input
										type="checkbox"
										checked={result.resolved}
										class="toggle toggle-success toggle-lg"
										on:click={()=>resolveAuction(result.id,result.resolved)}
										disabled={result.resolved}
									/>
								</td>
								<td>
									{#if result.expand?.resolvedBy?.name}
										<div class="badge badge-accent badge-lg">{result.expand?.resolvedBy?.name}</div>
									{:else}
										<span class="opacity-50 text-sm">Pending</span>
									{/if}
								</td>
								<td>
									{#if result.expand?.auction?.expand?.winner?.name}
										<div class="badge badge-primary badge-lg font-semibold gap-2">
											<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor">
												<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
											</svg>
											{result.expand?.auction?.expand?.winner?.name}
										</div>
									{:else}
										<span class="opacity-50 text-sm">No winner</span>
									{/if}
								</td>
								<td class="text-sm opacity-70">{new Date(result.created).toLocaleString()}</td>
							</tr>
						{/each}
					{:else}
						<tr>
							<td colspan="6" class="text-center py-12">
								<div class="flex flex-col items-center gap-3 opacity-50">
									<svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
									</svg>
									<p class="text-lg">No auction results found</p>
								</div>
							</td>
						</tr>
					{/if}
				</tbody>
			</table>
		</div>
	</div>
</div>
<AuthGuard requiredRole="manager" />