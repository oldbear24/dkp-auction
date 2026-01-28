<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { writable } from 'svelte/store';
	import pb from '$lib/pocketbase';
	import type { RecordModel } from 'pocketbase';
	import AuthGuard from '../../components/AuthGuard.svelte';
	import { user  as curentUser } from '$lib/stores/store';
	let users = writable<RecordModel[]>([]);
	let searchQuery = writable('');
	let verifiedFilter = writable('');
	let selectedUsers = writable<string[]>([]);
	let showDialog = writable(false);
	let tokensAmmount = writable(0);
	let tokensChangeReason = writable('');

	$: allSelected = $users.length === $selectedUsers.length;
	const userFields = 'id,name,avatar,tokens,validated,collectionId,discordId';
	async function fetchUsers() {
		try {
			let query = '';
			if($searchQuery!=''){
				query+=`name ~ '${$searchQuery}'`
			}
			if (query != '' && $verifiedFilter!='') {
				query += ' && ';
			}
			if($verifiedFilter!='')
			{
				query+='validated = '+ $verifiedFilter
			}
			console.debug(query)
			selectedUsers.set([]);
			const response = await pb.collection('users').getFullList({
				filter: query ? `${query}` : '',
				sort: '+name:lower',
				fields: userFields,
				requestKey:"get_users_management"
				
			});
			users.set(response);
		} catch (error) {
			console.error('Error fetching users:', error);
		}
	}

	onMount(() => {
		pb.realtime.subscribe('manage_users_update', userUpdateCallback);
		fetchUsers();
	});
	onDestroy(() => {
		pb.realtime.unsubscribe('manage_users_update');
	});

	let updateQueue: string[] = [];
	function userUpdateCallback(data: any) {
		updateQueue.push(data.userId);
	}

	function updateUserInTableWorker() {
		if (updateQueue.length > 0) {
			let id = updateQueue.shift();
			if (id) {
				console.debug('User update', id, updateQueue.length),
					pb
						.collection('users')
						.getOne(id, { fields: userFields })
						.then((rec) => {
							users.update((currentItems) => {
								return currentItems.map((item) =>
									item.id === rec.id ? { ...item, ...rec } : item
								);
							});
						});
			}
		}
	}
	setInterval(updateUserInTableWorker, 1);

	searchQuery.subscribe((query) => {
		fetchUsers();
	});
	verifiedFilter.subscribe((filter) => {
		fetchUsers();
	});
	function toggleAllCheckboxes() {
		console.debug('All checked:', allSelected);

		if (allSelected) {
			selectedUsers.set([]);
		} else {
			selectedUsers.set([...$users.map((user) => user.id)]);
		}

		console.debug('Selected users:', $selectedUsers);
	}

	async function changeValidation(userId: string, validated: boolean) {
		pb.send(`/api/set-validated/${userId}`, {
			method: 'POST',
			body: JSON.stringify({ validated: !validated })
		}).catch((err) => {
			console.error(err);
		});
	}

	function openDialog() {
		if ($selectedUsers.length > 0) {
			showDialog.set(true);
		}
	}

	function closeDialog() {
		showDialog.set(false);
		tokensAmmount.set(0);
		tokensChangeReason.set('');
	}

	async function changeTokens() {
		try {
			const userIds = $selectedUsers;
			const newTokens = $tokensAmmount;
			const changeReason = $tokensChangeReason;
			await pb.send('/api/change-tokens', {
				method: 'POST',
				body: JSON.stringify({ userIds: userIds, amount: newTokens,reason: changeReason })
			});

			closeDialog();
			selectedUsers.set([]);
		} catch (error) {
			console.error('Error changing tokens:', error);
		}
	}

	function openFileDialog() {
		const input = document.createElement('input');
		input.type = 'file';
		input.accept = '.csv';
		input.onchange = async (event: Event) => {
			const target = event.target as HTMLInputElement;
			if (target.files && target.files.length > 0) {
				const file = target.files[0];
				file.text().then(async (text) => {
					const userIds = text.split("\n").map(id => id.trim());	
					console.log(userIds)			
					selectedUsers.set($users.map(user => user.id));
					selectedUsers.set($users.filter(user => userIds.includes(user.discordId)).map(user => user.id));
					console.log($selectedUsers)
				});
			}
		};
		input.click();
	}

	function filterVerified(filter: string) {
		verifiedFilter.set(filter);
	}
	function deleteUser(id: string,name: string){
		if(!confirm("Do you realy want to delete user "+name))return
		pb.collection("users").delete(id).then((x=>{
			users.update(x=>x.filter(u=>u.id!=id))
		})).catch(x=>console.error("Could not delete user",id,x))
		}
		function clearTokens(){
			let percString = prompt("How many percent do you want to remove.","0")
			if(percString==null||percString=="0")return
			let perc = parseInt(percString)
			pb.send("api/clear-tokens",{method:"POST",body:JSON.stringify({percentage:perc})})		
		}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
		<h1 class="text-2xl font-bold mb-3 flex items-center gap-2">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
			</svg>
			Manage Users
		</h1>
		
		<!-- Stats -->
		<div class="stats stats-vertical lg:stats-horizontal shadow-xl mt-4 border border-base-content/10 w-full">
			<div class="stat bg-primary/10">
				<div class="stat-figure text-primary">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
					</svg>
				</div>
				<div class="stat-title">Total Users</div>
				<div class="stat-value text-primary">{$users.length}</div>
			</div>
			<div class="stat bg-accent/10">
				<div class="stat-figure text-accent">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 7h6m0 10v-3m-3 3h.01M9 17h.01M9 14h.01M12 14h.01M15 11h.01M12 11h.01M9 11h.01M7 21h10a2 2 0 002-2V5a2 2 0 00-2-2H7a2 2 0 00-2 2v14a2 2 0 002 2z" />
					</svg>
				</div>
				<div class="stat-title">Average Tokens</div>
				<div class="stat-value text-accent">{($users.reduce((a,n)=>a+n.tokens,0)/$users.length).toFixed(2)}</div>
			</div>
			<div class="stat bg-success/10">
				<div class="stat-figure text-success">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
				</div>
				<div class="stat-title">Total Tokens</div>
				<div class="stat-value text-success">{$users.reduce((a,n)=>a+n.tokens,0)}</div>
			</div>
		</div>
	</div>

	<!-- Search and Actions -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10 space-y-4">
		<div class="form-control">
			<div class="input-group">
				<input
					type="text"
					placeholder="Search by name..."
					class="input input-bordered w-full shadow-sm"
					bind:value={$searchQuery}
				/>
				<button class="btn btn-square btn-primary" aria-label="Search">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
					</svg>
				</button>
			</div>
		</div>

		<div class="flex flex-wrap gap-3">
			<button class="btn btn-primary gap-2" on:click={openDialog} disabled={$selectedUsers.length === 0}>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Change Tokens ({$selectedUsers.length})
			</button>
			<button class="btn btn-secondary gap-2" on:click={openFileDialog}>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12" />
				</svg>
				Import Users
			</button>
			<button class="btn btn-error gap-2" on:click={clearTokens}>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
				</svg>
				Clear Tokens
			</button>
		</div>

		<!-- Filter Tabs -->
		<div class="tabs tabs-boxed bg-base-100 shadow-inner">
			<button 
				class="tab gap-2 {$verifiedFilter === '' ? 'tab-active' : ''}" 
				on:click={() => filterVerified('')}
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
				</svg>
				All Users
			</button>
			<button 
				class="tab gap-2 {$verifiedFilter === 'true' ? 'tab-active' : ''}" 
				on:click={() => filterVerified('true')}
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Validated
			</button>
			<button 
				class="tab gap-2 {$verifiedFilter === 'false' ? 'tab-active' : ''}" 
				on:click={() => filterVerified('false')}
			>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Unvalidated
			</button>
		</div>
	</div>
	<!-- Users Table -->
	<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
		<div class="overflow-x-auto rounded-box border border-base-content/10 bg-base-100 shadow-inner">
			<table class="table table-zebra">
				<thead class="bg-base-300">
					<tr>
						<th class="text-base">
							<label>
								<input
									type="checkbox"
									class="checkbox checkbox-primary"
									bind:checked={allSelected}
									on:click={toggleAllCheckboxes}
								/>
							</label>
						</th>
						<th class="text-base">User</th>
						<th class="text-base">Tokens</th>
						<th class="text-base">Status</th>
						<th class="text-base">Actions</th>
					</tr>
				</thead>
				<tbody>
					{#each $users as user}
						<tr class="hover">
							<th>
								<label>
									<input
										type="checkbox"
										bind:group={$selectedUsers}
										name={user.id}
										value={user.id}
										class="checkbox checkbox-primary user-checkbox"
									/>
								</label>
							</th>
							<td>
								<div class="flex items-center gap-3">
									<div class="avatar avatar-placeholder">
										<div class="w-12 rounded-full ring-primary ring-offset-base-100 ring-2 ring-offset-2">
											{#if user.avatar}
											<img
												src={pb.files.getURL(user, user.avatar, { thumb: '100x100' })}
												alt="{user.name} profile image"
											/>
											{:else}
											<span class="text-xl font-bold">{user.name.substring(0,2)}</span>
											{/if}
										</div>
									</div>
									<div>
										<div class="font-bold">{user.name}</div>
									</div>
								</div>
							</td>
							<td>
								<div class="badge badge-primary badge-lg font-semibold">{user.tokens}</div>
							</td>
							<td>
								<input
									type="checkbox"
									on:click={() => changeValidation(user.id, user.validated)}
									bind:checked={user.validated}
									class="toggle toggle-success"
								/>
							</td>
							<td>
								<button 
									class="btn btn-error btn-sm gap-2" 
									disabled={user.id==$curentUser?.id} 
									on:click={()=>deleteUser(user.id,user.name)}
								>
									<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
									</svg>
									Delete
								</button>
							</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	</div>
</div>

{#if $showDialog}
	<div class="modal modal-open">
		<div class="modal-box max-w-lg shadow-2xl">
			<h2 class="text-xl font-bold mb-4 flex items-center gap-2">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Change Tokens
			</h2>
			<div class="bg-base-200 rounded-box p-4 mb-4">
				<p class="text-sm opacity-70">Selected Users</p>
				<p class="text-2xl font-bold">{$selectedUsers.length}</p>
			</div>
			<div class="form-control mb-4">
				<label class="label" for="tokensAmmount">
					<span class="label-text font-semibold">Token Amount</span>
				</label>
				<input
					id="tokensAmmount"
					type="number"
					class="input input-bordered w-full"
					bind:value={$tokensAmmount}
					placeholder="Enter amount..."
				/>
			</div>
			<div class="form-control mb-4">
				<label class="label" for="tokensReason">
					<span class="label-text font-semibold">Reason</span>
				</label>
				<input
					id="tokensReason"
					type="text"
					class="input input-bordered w-full"
					bind:value={$tokensChangeReason}
					placeholder="Enter reason..."
				/>
			</div>
			<div class="flex gap-2 justify-end">
				<button class="btn gap-2" on:click={closeDialog}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					Cancel
				</button>
				<button class="btn btn-primary gap-2" on:click={changeTokens}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7" />
					</svg>
					Submit
				</button>
			</div>
		</div>
	</div>
{/if}
<AuthGuard requiredRole="manager" />