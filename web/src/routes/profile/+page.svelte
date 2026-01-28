<script lang="ts">
	import { goto } from '$app/navigation';
	import pb from '$lib/pocketbase';
	import { user } from '$lib/stores/store';
	import type { ListResult, RecordModel } from 'pocketbase';
	import AuthGuard from '../../components/AuthGuard.svelte';
	import { onMount } from 'svelte';
	let deleteToggle = false;
	let transactions: ListResult<RecordModel> = {
		page: 1,
		perPage: 10,
		totalItems: 0,
		totalPages: 0,
		items: []
	};
	let currentPage = 1;
	const itemsPerPage = 10;

	function getHumanReadableRole(role: string) {
		switch (role) {
			case 'admin':
				return 'Admin';
			case 'member':
				return 'Member';
			case 'manager':
				return 'Manager';
			case 'lootCouncil':
				return 'Loot council';
			default:
				return role;
		}
	}

	function deleteUser() {
		if ($user != null && confirm("Do you really want to delete your account")) {
			pb.collection("users").delete($user.id).then((x => {
				pb.authStore.clear();
				user.set(null);
				goto('/');
			})).catch(x => console.error("Could not delete account ERROR:", x));
		}
	}

	function getTransactions(page = 1) {
		console.log('Fetching transactions for page:', page);
		currentPage = page;
		pb.collection('transactions')
			.getList(page, itemsPerPage, { filter: 'user = "' + $user?.id + '"' ,fields: 'amount,note,created',sort: '-created' })
			.then((resp) => {
				transactions = resp;
				console.debug('Transactions:', transactions.items);
			})
			.catch((error) => {
				console.error('Error fetching transactions:', error);
			});
	}
	onMount(() => {
		getTransactions()
	});
	$: if ($user) {
		getTransactions(); // Call getTransactions when $user is available
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
		<h1 class="text-3xl font-bold flex items-center gap-3">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-8 w-8" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
			</svg>
			User Profile
		</h1>
	</div>

	<!-- Profile Info Card -->
	<div class="stats stats-vertical lg:stats-horizontal shadow-xl border border-base-content/10 w-full">
		<div class="stat">
			<div class="flex items-center gap-4">
				<div class="avatar avatar-placeholder">
					<div class="w-24 rounded-full ring-primary ring-offset-base-100 ring-4 ring-offset-2 shadow-lg">
						{#if $user?.avatar}
						<img
							src={pb.files.getURL($user, $user?.avatar)}
							alt="{$user?.name} profile image"
						/>
						{:else}
						<span class="text-3xl font-bold">{$user?.name.substring(0,2)}</span>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<div class="stat">
			<div class="stat-title flex items-center gap-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
				</svg>
				Username
			</div>
			<div class="stat-value text-2xl">{$user?.name}</div>
		</div>

		<div class="stat">
			<div class="stat-title flex items-center gap-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
				</svg>
				Roles
			</div>
			<div class="flex flex-wrap gap-2 mt-2">
				{#each $user?.role as r}
					<div class="badge badge-accent badge-lg font-semibold">{getHumanReadableRole(r)}</div>
				{/each}
			</div>
		</div>
	</div>

	<!-- Token Stats -->
	<div class="stats stats-vertical lg:stats-horizontal shadow-xl border border-base-content/10 w-full">
		<div class="stat bg-accent/10">
			<div class="stat-figure text-accent">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<div class="stat-title">Total Tokens</div>
			<div class="stat-value text-accent">{$user?.tokens}</div>
			<div class="stat-desc">Your total token balance</div>
		</div>

		<div class="stat bg-error/10">
			<div class="stat-figure text-error">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
				</svg>
			</div>
			<div class="stat-title">Reserved Tokens</div>
			<div class="stat-value text-error">{$user?.reservedTokens}</div>
			<div class="stat-desc">Locked in active bids</div>
		</div>

		<div class="stat bg-info/10">
			<div class="stat-figure text-info">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-10 w-10" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
			</div>
			<div class="stat-title">Usable Tokens</div>
			<div class="stat-value text-info">{$user?.tokens-$user?.reservedTokens}</div>
			<div class="stat-desc">Available to bid</div>
		</div>
	</div>

	<!-- Transactions Section -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
		<h2 class="text-2xl font-bold mb-4 flex items-center gap-3">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
			</svg>
			Transaction History
		</h2>
		<div class="overflow-x-auto rounded-box border border-base-content/10 bg-base-100 shadow-inner">
			<table class="table table-zebra">
				<thead class="bg-base-300">
					<tr>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
								</svg>
								Amount
							</div>
						</th>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
								</svg>
								Note
							</div>
						</th>
						<th class="text-base">
							<div class="flex items-center gap-2">
								<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
									<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
								</svg>
								Date
							</div>
						</th>
					</tr>
				</thead>
				<tbody>
					{#if transactions.items.length > 0}
						{#each transactions.items as transaction}
						<tr class="hover">
							<td>
								<div class="font-semibold {transaction.amount > 0 ? 'text-success' : 'text-error'}">
									{transaction.amount > 0 ? '+' : ''}{transaction.amount}
								</div>
							</td>
							<td class="max-w-md truncate">{transaction.note}</td>
							<td class="text-sm opacity-70">{new Date(transaction.created).toLocaleString()}</td>
						</tr>
						{/each}
					{:else}
						<tr>
							<td colspan="3" class="text-center py-8">
								<div class="flex flex-col items-center gap-2 opacity-50">
									<svg xmlns="http://www.w3.org/2000/svg" class="h-12 w-12" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
									</svg>
									<p>No transactions yet</p>
								</div>
							</td>
						</tr>
					{/if}
				</tbody>
			</table>
		</div>
		{#if transactions.totalPages > 1}
			<div class="flex justify-center mt-4">
				<div class="join shadow-lg">
					<button class="join-item btn" on:click={() => getTransactions(currentPage - 1)} disabled={currentPage === 1} aria-label="Previous page">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
						</svg>
					</button>
					<button class="join-item btn pointer-events-none">Page {currentPage} of {transactions.totalPages}</button>
					<button class="join-item btn" on:click={() => getTransactions(currentPage + 1)} disabled={currentPage === transactions.totalPages} aria-label="Next page">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
						</svg>
					</button>
				</div>
			</div>
		{/if}
	</div>

	<!-- Danger Zone -->
	<div class="bg-error/10 rounded-box p-6 shadow-lg border-2 border-error/30">
		<h2 class="text-2xl font-bold mb-4 text-error flex items-center gap-3">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
			</svg>
			Danger Zone
		</h2>
		<p class="text-sm opacity-80 mb-4">This action is permanent and cannot be undone</p>
		<div class="flex items-center gap-4">
			<label class="cursor-pointer flex items-center gap-2">
				<input type="checkbox" class="toggle toggle-error" bind:checked={deleteToggle} />
				<span class="text-sm font-semibold">I understand the consequences</span>
			</label>
			<button 
				disabled={!deleteToggle} 
				on:click={deleteUser} 
				class="btn btn-error gap-2"
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
				</svg>
				Delete My Account
			</button>
		</div>
	</div>
</div>
<AuthGuard />