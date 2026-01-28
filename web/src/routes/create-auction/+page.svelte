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

<div class="flex justify-center">
	<div class="w-full max-w-2xl space-y-6">
		<!-- Page Header -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h1 class="text-3xl font-bold flex items-center gap-3">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
				</svg>
				Create New Auction
			</h1>
			<p class="text-sm opacity-70 mt-2">Fill in the details to list a new item for auction</p>
		</div>

		<!-- Form Card -->
		<form on:submit={createAuction} class="bg-base-200 rounded-box p-8 shadow-xl border border-base-content/10 space-y-6">
			
			<!-- Item Name -->
			<div class="form-control">
				<label class="label" for="itemName">
					<span class="label-text text-lg font-semibold flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
						</svg>
						Item Name
					</span>
				</label>
				<input 
					required 
					id="itemName" 
					type="text" 
					class="input input-bordered input-lg w-full shadow-sm" 
					bind:value={$itemName} 
					list="my-options"
					placeholder="Enter item name..."
				/>
			</div>

			<!-- Rarity -->
			<div class="form-control">
				<label class="label" for="rarity">
					<span class="label-text text-lg font-semibold flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z" />
						</svg>
						Rarity
					</span>
				</label>
				<select id="rarity" class="select select-bordered select-lg w-full shadow-sm" bind:value={$rarity}>
					{#each RARITY_OPTIONS as opt}
						<option value={opt} class={RARITY_MAP[opt].class}>{RARITY_MAP[opt].label}</option>
					{/each}
				</select>
			</div>

			<!-- Description -->
			<div class="form-control">
				<label class="label" for="description">
					<span class="label-text text-lg font-semibold flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h7" />
						</svg>
						Description
					</span>
				</label>
				<textarea 
					id="description" 
					class="textarea textarea-bordered textarea-lg w-full h-32 shadow-sm" 
					bind:value={$description}
					placeholder="Describe the item..."
				></textarea>
			</div>

			<!-- Starting Bid -->
			<div class="form-control">
				<label class="label" for="startingBid">
					<span class="label-text text-lg font-semibold flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						Starting Bid
					</span>
				</label>
				<input 
					id="startingBid" 
					type="number" 
					class="input input-bordered input-lg w-full shadow-sm" 
					bind:value={$startingBid}
					placeholder="Enter starting bid amount..."
				/>
			</div>

			<!-- Main Image -->
			<div class="form-control">
				<label class="label" for="mainImage">
					<span class="label-text text-lg font-semibold flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						Main Image
					</span>
				</label>
				<input 
					id="mainImage" 
					class="file-input file-input-bordered file-input-lg w-full shadow-sm" 
					type="file" 
					on:change={handleFileChange}
					accept="image/*"
				/>
			</div>

			<!-- End Time -->
			<div class="form-control">
				<label class="label" for="endTime">
					<span class="label-text text-lg font-semibold flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
						</svg>
						End Time
					</span>
				</label>
				<input 
					required 
					id="endTime" 
					class="input input-bordered input-lg w-full shadow-sm" 
					bind:value={$endTime} 
					type="datetime-local"
				/>
				<div class="label-text-alt">Quick select:</div>
				<div class="flex gap-2">
					<button class="btn btn-sm btn-outline gap-2" on:click={()=>getNextDate(1)} type="button">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						+1 Day
					</button>
					<button class="btn btn-sm btn-outline gap-2" on:click={()=>getNextDate(2)} type="button">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						+2 Days
					</button>
					<button class="btn btn-sm btn-outline gap-2" on:click={()=>getNextDate(7)} type="button">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						+1 Week
					</button>
				</div>
			</div>

			<!-- Submit Button -->
			<div class="form-control pt-4">
				<button class="btn btn-primary btn-lg w-full gap-2 shadow-lg" type="submit">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					Create Auction
				</button>
			</div>
		</form>
	</div>
</div>
<datalist id="my-options">
	{#each $items as item}
		<option value={item}></option>
	{/each}
  </datalist>
<AuthGuard requiredRole="manager" />