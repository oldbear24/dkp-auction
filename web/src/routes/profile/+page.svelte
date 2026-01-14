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

<div class="flex w-full flex-col">
	<div class="stats stats-vertical lg:stats-horizontal shadow">
		<div class="stat">
			<div class="flex items-center gap-3">
				<div class="avatar avatar-placeholder">
					<div class="w-18 rounded-full ring-primary ring-offset-base-100 ring ring-offset-1">
						{#if $user?.avatar}
						<img
							src={pb.files.getURL($user, $user?.avatar)}
							alt="{$user?.name} profile image"
						/>
						{:else}
						<span class="text-xl">{$user?.name.substring(0,2)}</span>
						{/if}
					</div>
				</div>
			</div>
		</div>

		<div class="stat">
			<div class="stat-title">Username</div>
			<div class="stat-value">{$user?.name}</div>
		</div>

		<div class="stat">
			<div class="stat-title">Roles</div>
			<div class="text-info flex flex-wrap">{#each $user?.role as r}
				<div class="badge badge-outline badge-accent m-1">{getHumanReadableRole(r)}</div>
				{/each}
			</div>
		</div>
	</div>
	<div class="divider"></div>
	<div class="stats stats-vertical lg:stats-horizontal shadow">
		<div class="stat">
			<div class="stat-title">Tokens</div>
			<div class="stat-value text-accent">{$user?.tokens}</div>
		</div>

		<div class="stat">
			<div class="stat-title">Reserved tokens</div>
			<div class="stat-value text-error">{$user?.reservedTokens}</div>
		</div>

		<div class="stat">
			<div class="stat-title">Usable tokens</div>
			<div class="stat-value text-info">{$user?.tokens-$user?.reservedTokens}</div>
		</div>
	</div>
	<div class="divider"></div>
	<h2 class="text-2xl font-bold pb-2">Transactions</h2>
	<div class="overflow-x-auto rounded-box border border-base-content/5 bg-base-100">
		<table class="table">
			<!-- head -->
			<thead>
				<tr>
					<th>Amount</th>
					<th>Note</th>
					<th>Created</th>
				</tr>
			</thead>
			<tbody>
				{#each transactions.items as transaction}
				<tr>
					<td>{transaction.amount}</td>
					<td>{transaction.note}</td>
					<td>{new Date(transaction.created).toLocaleString()}</td>
				</tr>
				{/each}
			</tbody>
		</table>
	</div>
	<div class="join justify-center pt-2">
		<button class="join-item btn" tabindex="-1" on:click={() => getTransactions(currentPage - 1)} disabled={currentPage === 1}>«</button>
		<button class="join-item btn" tabindex="-1">Page {currentPage} of {transactions.totalPages}</button>
		<button class="join-item btn" tabindex="-1" on:click={() => getTransactions(currentPage + 1)} disabled={currentPage === transactions.totalPages}>»</button>
	</div>
	<div class="divider"></div>

	<div>
		<input type="checkbox" class="toggle toggle-error" bind:checked={deleteToggle} />
		<button disabled={!deleteToggle} on:click={deleteUser} class="btn btn-error w-45">Delete my account</button>
	</div>
</div>
<AuthGuard />