<script lang="ts">
  import type { ListResult, RecordModel } from 'pocketbase';
  import pb, { subscribeToAuctionUpdate, unsubscribeFromAuctionUpdates } from '../lib/pocketbase';
  import AuctionItem from '../components/AuctionItem.svelte';
  import { user } from '$lib/stores/store';
  const itemsPerPage = 9;
  let items: ListResult<RecordModel> = { page: 1, perPage: itemsPerPage, totalItems: 0, totalPages: 0, items: [] };
  let currentPage = 1;
  let searchQuery = '';
  async function fetchItems(page: number) {
    const date = new Date();
        date.setDate(date.getDate() -2 );
    let filterString = `endTime > '${new Date(date).toISOString()}'`;
    
    if (searchQuery != '') {
      filterString += ` && itemName ~ '${searchQuery}%'`;
    }
    console.debug('Fetching items with filter:', filterString);
    const records = await pb.collection('auctions').getList(page, itemsPerPage, { sort: "-endTime", filter: filterString });
    items = records;
  }

  function updateItem(record: RecordModel) {
    const index = items.items.findIndex(item => item.id === record.id);
    console.debug('Updating item:', record.id, index);

    if (index !== -1) {
      items.items[index]=record;
      items.items = [...items.items]; // Trigger reactivity
    }
  }

  async function subscribeToCurrentPage() {
    await unsubscribeFromAuctionUpdates();
    console.debug('Subscribing to updates for current page');
    const recordIds = items.items.map(item => item.id);
    recordIds.forEach(recordId => subscribeToAuctionUpdate(recordId, updateItem));
  }

  function changePage(page: number) {
    currentPage = page;
    fetchItems(currentPage).then(subscribeToCurrentPage);
  }

  function handleSearch() {
    fetchItems(1).then(subscribeToCurrentPage);
  }

  fetchItems(currentPage).then(subscribeToCurrentPage);
</script>
{#if $user}
<div class="container mx-auto px-4 overflow-x-auto pb-4">
  <div class="flex justify-between items-center mb-4">
    <input 
      type="text" 
      placeholder="Search auctions..." 
      class="input input-bordered w-full max-w-xs" 
      bind:value={searchQuery}
      on:input={handleSearch} 
    />
  </div>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
    {#each items.items as item}
      <AuctionItem {item} />
    {/each}
  </div> 

</div>
<div class="join flex justify-center pt-2">
  <button class="join-item btn" tabindex="-1" on:click={() => changePage(currentPage - 1)} disabled={currentPage === 1}>«</button>
  <button class="join-item btn" tabindex="-1">Page {currentPage} of {items.totalPages}</button>
  <button class="join-item btn" tabindex="-1" on:click={() => changePage(currentPage + 1)} disabled={currentPage === items.totalPages||items.totalPages===0}>»</button>
</div>
{:else}
<div class="flex justify-center items-center h-screen">
  <p class="text-xl">You are not logged in. Please log in to view the auction items.</p>
</div>
{/if}