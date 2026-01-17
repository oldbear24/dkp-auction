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
  let isFavourite = false;
  let favouriteId: string | null = null;

$: {
  if (!item || !$user) {
    isFavourite = false;
    favouriteId = null;
  } else {
    const favs = item.favourites_via_auction || item.expand?.favourites_via_auction || [];
    // PocketBase may return relation fields either as an ID string or as an expanded record.
    // Normalize to a single userId value before comparing with the current user.
    const fav = favs.find((f: any) => {
      const favUserId = typeof f?.user === 'string' ? f.user : f?.user?.id;
      return favUserId === $user?.id;
    });
    isFavourite = !!fav;
    favouriteId = fav?.id ?? null;
  }
}

async function toggleFavourite(event?: Event) {
  event?.stopPropagation?.();
  if (!$user) {
    showToast('Please log in to manage favourites', 'error');
    return;
  }
  try {
    if (isFavourite && favouriteId) {
      await pb.collection('favourites').delete(favouriteId);
      showToast('Removed from favourites', 'success');
      isFavourite = false;
      favouriteId = null;
    } else {
      const rec = await pb.collection('favourites').create({ auction: item.id, user: $user.id });
      showToast('Added to favourites', 'success');
      isFavourite = true;
      favouriteId = rec.id;
    }
  } catch (err:any) {
    showToast(err?.message || 'Action failed', 'error');
  }
}
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
</script>

<div class="card bg-base-200  {item.winner==$user?.id?'ring-2 shadow-[0_0_15px] ring-accent shadow-accent':'shadow-lg'}	">
  <figure>
    <img src={getImage(item)} alt={item.itemName} class="w-full h-48 object-cover rounded-lg" />
  </figure>
  <div class="card-body">
    <div class="flex justify-between items-start w-full">
      <h2 class="card-title font-bold underline text-xl decoration-gray-300">{item.itemName}</h2>
      <button class="btn btn-ghost btn-circle" on:click|stopPropagation={toggleFavourite} aria-label="Toggle favourite">
        {#if isFavourite}
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-red-500" viewBox="0 0 24 24" fill="currentColor"><path d="M12 21.35l-1.45-1.32C5.4 15.36 2 12.28 2 8.5 2 6 4 4 6.5 4c1.74 0 3.41.81 4.5 2.09C12.09 4.81 13.76 4 15.5 4 18 4 20 6 20 8.5c0 3.78-3.4 6.86-8.55 11.54L12 21.35z"/></svg>
        {:else}
          <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.84 4.61a5.5 5.5 0 00-7.78 0L12 5.67l-1.06-1.06a5.5 5.5 0 10-7.78 7.78L12 21.23l8.84-8.84a5.5 5.5 0 000-7.78z"/></svg>
        {/if}
      </button>
    </div>
    <p><RarityLabel value={item.rarity} asBadge /></p>
    <p class="whitespace-pre-wrap overflow-auto max-h-20">{item.description}</p>
    <p class="font-bold text-lg">Current Bid: {currentBid}</p>
    <div class="flex space-x-2">
      {#if $countdown.years > 0}
        <div class="countdown">
          <span style="--value:{$countdown.years};"></span>y
        </div>
      {/if}
      {#if $countdown.months > 0}
        <div class="countdown">
          <span style="--value:{$countdown.months};"></span>m
        </div>
      {/if}
      {#if $countdown.days > 0}
        <div class="countdown">
          <span style="--value:{$countdown.days};"></span>d
        </div>
      {/if}
      {#if $countdown.hours > 0}
        <div class="countdown">
          <span style="--value:{$countdown.hours};"></span>h
        </div>
      {/if}
      {#if $countdown.minutes > 0}
        <div class="countdown">
          <span style="--value:{$countdown.minutes};"></span>m
        </div>
      {/if}
      {#if $countdown.seconds > 0}
        <div class="countdown">
          <span style="--value:{$countdown.seconds};"></span>s
        </div>
      {/if}
    </div>
    
    {#if item.state === 'ongoing'}
      <button class="btn btn-primary" on:click={openModal}>Place Bid</button>
    {:else if item.state === 'pending'}
      <div class="text-warning">Auction not started</div>
      <button class="btn btn-disabled">Waiting to Start</button>
    {:else}
      <div class="text-error">Auction ended</div>
      <button class="btn btn-disabled">Completed</button>
    {/if}
  </div>
</div>

{#if $showModal && item.state === 'ongoing'}
  <div class="modal modal-open">
    <div class="modal-box">
      <h2 class="card-title">{item.itemName}</h2>
      <figure>
        <img src={getImage(item)} alt={item.itemName} class="w-full h-48 object-cover rounded-lg" />
      </figure>
      <p>{item.description}</p>
      <p>Current Bid: {currentBid}</p>
      <div class="form-control">
        <label class="label" for="bidAmount">
          <span class="label-text">Your Bid Amount</span>
        </label>
        <input type="number" id="bidAmount" class="input input-bordered" bind:value={$bidAmount} min={item.bid} />
      </div>
      <div class="modal-action">
        <button class="btn btn-primary" on:click={confirmBid}>Confirm Bid</button>
        <button class="btn" on:click={closeModal}>Close</button>
      </div>
    </div>
  </div>
{/if}