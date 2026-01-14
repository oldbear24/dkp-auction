<script lang="ts">
	import {  writable } from 'svelte/store';
	import pb from '$lib/pocketbase';
	import AuthGuard from '../../components/AuthGuard.svelte';
	let items = writable<string[]>([])
	let itemName = writable('');
	let description = writable('');
	let mainImage = writable<File | null>(null);
	// Rarity for auction items (shared source)
	import { RARITY_OPTIONS, RARITY_MAP, type RarityValue } from '$lib/rarity';
	let rarity = writable<RarityValue>();
	let endTime = writable<string>(""); // Set default value
        let startingBid = writable(0);

    function getNextDate(days: number) {
        const date = new Date();
        date.setDate(date.getDate() + days);
        date.setMinutes(date.getMinutes() - date.getTimezoneOffset());
        date.setMilliseconds(0);
        date.setSeconds(0);
        endTime.set(new Date(date).toISOString().slice(0, -1));
    }
	async function createAuction() {
		try {
			const data = {
				itemName: $itemName,
				rarity: $rarity,
				description: $description,
				startingBid: $startingBid,
				endTime: new Date($endTime).toISOString(),
                mainImage: $mainImage
			};
			console.debug('Creating auction:', data);
			const response = await pb.collection('auctions').create(data);
			console.debug('Auction created:', await response);
			// Reset form fields
			itemName.set('');
			description.set('');
			mainImage.set(null);
			endTime.set("");
		} catch (error) {
			console.error('Error:', error);
		}
	}
    
	function handleFileChange(event: Event) {
		const target = event.target as HTMLInputElement;
		if (target.files && target.files.length > 0) {
			mainImage.set(target.files[0]);
		}
	}
	function loadItems() {
		pb.collection('items').getFullList().then((data) => {
			items.set(data.map(item => item.name));
		}).catch((error) => {
			console.error('Error loading items:', error);
		});
	}
	loadItems();
</script>

<div class="justify-center flex w-full">
<form on:submit={createAuction}>
<fieldset class="fieldset bg-base-200 border-base-300 rounded-box w-xs border p-4">

	<legend class="fieldset-legend">Add item to auction</legend>
	<label class="fieldset-label" for="itemName">Item name</label>
	<input required id="itemName" type="text" class="input validator" bind:value={$itemName} list="my-options" />
	<label class="fieldset-label" for="rarity">Rarity</label>
	<select id="rarity" class="select" bind:value={$rarity}>
		{#each RARITY_OPTIONS as opt}
			<option value={opt} class={RARITY_MAP[opt].class}>{RARITY_MAP[opt].label}</option>
		{/each}
	</select>
	<label class="fieldset-label" for="description">Description</label>
	<textarea id="description" class="input min-h-40" bind:value={$description}></textarea>
	<label class="fieldset-label" for="startingBid">Starting bud</label>
	<input id="startingBid" type="number" class="input" bind:value={$startingBid} />
    <label class="fieldset-label" for="mainImage">Main Image</label>
    <input id="mainImage" class="file-input" type="file" on:change={handleFileChange} />
	<label class="fieldset-label" for="endTime">End Time</label>
	<input required id="endTime" class="input validator" bind:value={$endTime} type="datetime-local" defaultValue={getNextDate} />
	<div>
		<button class="btn btn-accent" on:click={()=>getNextDate(1)} type="button" >+1D</button>
		<button class="btn btn-accent" on:click={()=>getNextDate(7)} type="button" >+7D</button>	
	</div>
		<button class="btn btn-primary" type="submit" >Create Auction</button>
</fieldset>
</form>
</div>
<datalist id="my-options">
	{#each $items as item}
		<option value={item}></option>
	{/each}
  </datalist>
<AuthGuard requiredRole="manager" />