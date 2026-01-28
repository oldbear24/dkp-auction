<script lang="ts">
  export const ssr = false;

  import pb from '$lib/pocketbase';
  import { writable } from 'svelte/store';
  import { onDestroy, onMount } from 'svelte';
  import { showToast,user } from '$lib/stores/store';
	import type { RecordModel } from 'pocketbase';
	import RarityLabel from './RarityLabel.svelte';
  
  $: currentBid = item ? Math.max(item.startingBid, item.currentBid) : 0;

  export let item: RecordModel;
 
  let showModal = writable(false);
  let bidAmount = writable(item.bid);
  let countdown = writable({ years: 0, months: 0, days: 0, hours: 0, minutes: 0, seconds: 0 });
  function openModal() {
    if (item.state !== 'ongoing') return;
    showModal.set(true);
  }

  function closeModal() {
    showModal.set(false);
  }

  async function confirmBid() {
    try {
      console.debug('Placing bid:', item);
      const response = await pb.send(`/api/bid/${item.id}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({ amount: $bidAmount })
      });
      // Update the item with the new bid amount
      //item.bid = $bidAmount;
      showToast('Bid placed successfully!', 'success');
    } catch (error:any) {
      showToast(error.message, 'error');
    } finally {
      closeModal();
    }
  }

  function updateCountdown() {
    const endTime = new Date(item.endTime).getTime();
    const now = new Date().getTime();
    const distance = endTime - now;

    if (distance < 0) {
      countdown.set({ years: 0, months: 0, days: 0, hours: 0, minutes: 0, seconds: 0 });
      return;
    }

    const years = Math.floor(distance / (1000 * 60 * 60 * 24 * 365));
    const months = Math.floor((distance % (1000 * 60 * 60 * 24 * 365)) / (1000 * 60 * 60 * 24 * 30));
    const days = Math.floor((distance % (1000 * 60 * 60 * 24 * 30)) / (1000 * 60 * 60 * 24));
    const hours = Math.floor((distance % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    const minutes = Math.floor((distance % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((distance % (1000 * 60)) / 1000);

    countdown.set({ years, months, days, hours, minutes, seconds });
  }

  onMount(() => {
    console.debug('Setting interval');
    console.debug('Item', item);
    updateCountdown();
    const interval = setInterval(updateCountdown, 1000);
    return () => {
      console.debug('Clearing interval');
      clearInterval(interval);}
  });
    function getImage(record: RecordModel) {
      let imgUrl
      if ( record.mainImage==null||record.mainImage==''){ 
        imgUrl = "/no_image_placeholder_dark.png";
      }else{
       imgUrl = pb.files.getURL(record, record.mainImage, { 'thumb': '500x200' });
      }
      return imgUrl;
    }
  onDestroy( () => {
        console.log("Date Component removed")
    });


	function toggleFavourite(event: MouseEvent & { currentTarget: EventTarget & HTMLButtonElement; }) {
		 if (item?.favourites?.includes($user?.id)) {
      // Remove from favourites
      pb.send(`/api/remove-from-favourites/${item.id}`, {
        method: 'POST'
      }).then(() => {
        // Update local state
        item.favourites = item.favourites.filter((userId: string) => userId !== $user?.id);
        showToast('Removed from favourites', 'info');
      }).catch((error) => {
        showToast(`Error removing from favourites: ${error.message}`, 'error');
      });
    } else {
      // Add to favourites
      pb.send(`/api/add-to-favourites/${item.id}`, {
        method: 'POST'
      }).then(() => {
        // Update local state
        item.favourites = [...(item.favourites || []), $user?.id];
        showToast('Added to favourites', 'success');
      }).catch((error) => {
        showToast(`Error adding to favourites: ${error.message}`, 'error');
      });
    }
	}
</script>

<div class="card bg-base-200 shadow-xl hover:shadow-2xl transition-all duration-300 {item.winner==$user?.id?'ring-4 shadow-[0_0_20px] ring-accent shadow-accent':'border border-base-content/10'}">
  <figure class="relative">
    <img src={getImage(item)} alt={item.itemName} class="w-full h-56 object-cover" />
		{#if item.winner==$user?.id}
			<div class="absolute top-4 left-4 badge badge-accent badge-lg gap-2 shadow-lg">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
					<path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
				</svg>
				Winner!
			</div>
		{/if}
		<button 
			class="absolute top-4 right-4 btn btn-circle btn-sm bg-base-100/80 backdrop-blur-sm border-0 hover:bg-base-100 hover:scale-110 transition-all shadow-lg" 
			on:click={toggleFavourite} 
			aria-label="Toggle favourite"
		>
			{#if item?.favourites?.includes($user?.id)}
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-error" viewBox="0 0 24 24" fill="currentColor">
					<path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 6 4 4 6.5 4c1.74 0 3.41.81 4.5 2.09C12.09 4.81 13.76 4 15.5 4 18 4 20 6 20 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/>
				</svg>
			{:else}
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-base-content/50" viewBox="0 0 24 24" fill="none" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 10-7.78 7.78L12 21.23l8.84-8.84a5.5 5.5 0 000-7.78z"/>
				</svg>
			{/if}
		</button>
  </figure>
  <div class="card-body p-6">
    <div class="flex justify-between items-start w-full mb-2">
      <h2 class="card-title font-bold text-2xl">{item.itemName}</h2>
    </div>
    <div class="mb-3">
			<RarityLabel value={item.rarity} asBadge />
		</div>
    <p class="text-sm opacity-80 whitespace-pre-wrap overflow-auto max-h-20 mb-4">{item.description}</p>
    
		<!-- Current Bid Display -->
		<div class="bg-base-300 rounded-box p-4 mb-4 border border-primary/20">
			<div class="flex flex-col gap-3">
				<div class="flex items-center justify-between">
					<span class="text-sm opacity-70">Current Bid</span>
					<div class="flex items-center gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
						<span class="text-2xl font-bold text-primary">{currentBid}</span>
					</div>
				</div>
				<!-- Countdown Timer - Inline -->
				<div class="flex items-center justify-between border-t border-base-content/10 pt-3">
					<span class="text-sm opacity-70">Time Remaining</span>
					<div class="flex items-center gap-1 font-mono text-sm font-semibold">
						{#if $countdown.years > 0}
							<span class="bg-base-100 px-2 py-1 rounded">{$countdown.years}y</span>
						{/if}
						{#if $countdown.months > 0}
							<span class="bg-base-100 px-2 py-1 rounded">{$countdown.months}mo</span>
						{/if}
						{#if $countdown.days > 0}
							<span class="bg-base-100 px-2 py-1 rounded">{$countdown.days}d</span>
						{/if}
						{#if $countdown.hours > 0}
							<span class="bg-base-100 px-2 py-1 rounded">{$countdown.hours}h</span>
						{/if}
						{#if $countdown.minutes > 0}
							<span class="bg-base-100 px-2 py-1 rounded">{$countdown.minutes}m</span>
						{/if}
						{#if $countdown.seconds > 0 || ($countdown.years === 0 && $countdown.months === 0 && $countdown.days === 0 && $countdown.hours === 0 && $countdown.minutes === 0)}
							<span class="bg-base-100 px-2 py-1 rounded">{$countdown.seconds}s</span>
						{/if}
					</div>
				</div>
			</div>
		</div>
    
    {#if item.state === 'ongoing'}
      <button class="btn btn-primary w-full gap-2 shadow-lg" on:click={openModal}>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
				</svg>
				Place Bid
			</button>
    {:else if item.state === 'pending'}
      <div class="alert alert-warning shadow-lg">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span>Auction not started yet</span>
			</div>
    {:else}
      <div class="alert alert-error shadow-lg">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				<span>Auction has ended</span>
			</div>
    {/if}
  </div>
</div>

{#if $showModal && item.state === 'ongoing'}
  <div class="modal modal-open">
    <div class="modal-box max-w-2xl shadow-2xl">
      <h2 class="text-2xl font-bold mb-4 flex items-center gap-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
				</svg>
				{item.itemName}
			</h2>
      <figure class="mb-4">
        <img src={getImage(item)} alt={item.itemName} class="w-full h-64 object-cover rounded-xl shadow-lg" />
      </figure>
      <div class="space-y-4">
				<div class="bg-base-200 rounded-box p-4">
					<p class="text-sm opacity-70 mb-2">Description</p>
					<p class="whitespace-pre-wrap">{item.description}</p>
				</div>
				<div class="bg-base-200 rounded-box p-4">
					<div class="flex items-center justify-between">
						<span class="text-sm opacity-70">Current Bid</span>
						<div class="flex items-center gap-2">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
							<span class="text-2xl font-bold text-primary">{currentBid}</span>
						</div>
					</div>
				</div>
				<div class="form-control">
					<label class="label" for="bidAmount">
						<span class="label-text font-semibold">Your Bid Amount</span>
					</label>
					<input 
						type="number" 
						id="bidAmount" 
						class="input input-bordered w-full shadow-sm" 
						bind:value={$bidAmount} 
						min={item.bid} 
						placeholder="Enter your bid..."
					/>
				</div>
			</div>
      <div class="modal-action gap-2">
        <button class="btn btn-primary gap-2" on:click={confirmBid}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					Confirm Bid
				</button>
        <button class="btn gap-2" on:click={closeModal}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					Cancel
				</button>
      </div>
    </div>
  </div>
{/if}