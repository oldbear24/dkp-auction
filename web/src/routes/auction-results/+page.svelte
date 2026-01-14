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

<div class="container mx-auto">
	<h1 class="mb-4 text-2xl font-bold">Auction Results</h1>
	<fieldset class="fieldset bg-base-100 border-base-300 rounded-box w-64 border p-4">
		<legend class="fieldset-legend">Filter options</legend>
		<label class="fieldset-label">
			<input
				type="checkbox"
				bind:checked={$showUnresolvedOnly}
				on:change={fetchAuctionResults}
				class="toggle toggle-info"
			/>
			Show Unresolved Only
		</label>
	</fieldset>
	<div class="h-250 overflow-x-auto">
		<table class="table-pin-cols table-pin-rows table">
			<!-- head -->
			<thead>
				<tr>
					<th>Auction</th>
					<th>Descriptin</th>
					<th>Resolved</th>
					<th>Resolved By</th>
					<th>Winner</th>
					<th>Created</th>
				</tr>
			</thead>
			<tbody>
				{#each $auctionResults as result}
					<tr>
						<td>{result.expand?.auction.itemName}</td>
						<td>{result.expand?.auction.description}</td>
						<td
							><input
								type="checkbox"
								checked={result.resolved}
								class="toggle text-error checked:text-success"
								on:click={()=>resolveAuction(result.id,result.resolved)}
								disabled={result.resolved}
							/></td
						>
						<td>{result.expand?.resolvedBy?.name}</td>
						<td>{result.expand?.auction?.expand?.winner?.name}</td>
						<td>{new Date(result.created).toLocaleString()}</td>
					</tr>
				{/each}
			</tbody>
			<!-- foot -->
			<tfoot> </tfoot>
		</table>
	</div>
</div>
<AuthGuard requiredRole="manager" />