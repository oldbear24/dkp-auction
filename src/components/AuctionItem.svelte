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
</script>

<div class="card bg-base-200  {item.winner==$user?.id?'ring-2 shadow-[0_0_15px] ring-accent shadow-accent':'shadow-lg'}	">
  <figure>
    <img src={getImage(item)} alt={item.itemName} class="w-full h-48 object-cover rounded-lg" />
  </figure>
  <div class="card-body">
    <h2 class="card-title font-bold underline text-xl decoration-gray-300">{item.itemName}</h2>
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