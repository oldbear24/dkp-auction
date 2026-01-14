<script lang="ts">
	import { user } from '$lib/stores/store';
	import { onMount } from 'svelte';
	import pb, { subscribeToUserUpdate, unsubscribeFromUserUpdates } from '$lib/pocketbase';
	import ToastManager from './ToastManager.svelte';
	import { writable } from 'svelte/store';
	import NotificationsList from './NotificationsList.svelte';
	let notificationListComponent: NotificationsList;
	let notificationCount = writable(0);
  const openNotification = writable(false);
	async function loginWithDiscord() {
		try {
			unsubscribeFromUserUpdates();
			const authData = await pb.collection('users').authWithOAuth2({ provider: 'discord' });
			user.set(authData.record);
			await subscribeToUserUpdate(authData.record.id, (updatedUser) => {
				user.set(updatedUser);
			});
			await subsbscibeToNotifications();
		} catch (error) {
			console.error('Login failed:', error);
		}
	}

	function logout() {
		pb.authStore.clear();
		user.set(null);
		unsubscribeFromUserUpdates();
	}

	onMount(async () => {
		await pb.collection('users').authRefresh();
		const currentUser = pb.authStore.record;
		if (currentUser) {
			console.debug('User already logged in:', currentUser);
			user.set(currentUser);
			await subscribeToUserUpdate(currentUser.id, (updatedUser) => {
				user.set(updatedUser);
			});
			await subsbscibeToNotifications();
		}
	});

	async function subsbscibeToNotifications() {
		let resp = await pb.collection('notifications').getList(0, 1, { fields: 'id' });
		console.debug('Notifications:', resp);
		notificationCount.set(resp.totalItems);
		pb.collection('notifications').subscribe('*', (notification) => {
			console.debug('Notification:', notification);
			if (notification.action === 'create') notificationCount.update((count) => count + 1);
		});
	}
</script>

<nav class="navbar bg-base-300 shadow-sm">
	<div class="flex-1">
		<a class="btn btn-ghost text-xl" href="/">Auction House</a>
	</div>

	<div class="flex-none pr-3">
		{#if $user}
			<div
				data-tip="Your current tokens and usable tokens."
				class="tooltip tooltip-left tooltip-info ml-2 pr-3"
			>
				<span>Tokens: {$user.tokens} ({$user.tokens - $user.reservedTokens})</span>
			</div>
			<div class="indicator">
				{#if $notificationCount > 99}
					<span class="indicator-item badge badge-primary badge-xs">99+</span>
				{:else if $notificationCount>0}
					<span class="indicator-item badge badge-primary badge-xs">{$notificationCount}</span>
				{/if}
			<div class="dropdown dropdown-end">
				<button class="btn btn-ghost btn-circle avatar avatar-placeholder" aria-label="User menu">
					<div class="ring-primary ring-offset-base-100 w-10 rounded-full ring ring-offset-1">
						{#if $user.avatar}
							<img
								src={pb.files.getURL($user, $user.avatar, { thumb: '100x100' })}
								alt="User Avatar"
							/>
						{:else}
							<span class="text-xl">{$user.name.substring(0, 2)}</span>
						{/if}
					</div>
				</button>
				<ul
					class="menu menu-compact dropdown-content bg-base-200 rounded-box mt-3 w-52 p-2 shadow"
					style="z-index: 100;"
				>
					<li><a href="/profile">Profile</a></li>
          <li><button  on:click="{() => {openNotification.set(true)}}">
            Notifications <div class="badge badge-sm">{$notificationCount}</div>
          </button></li>
					{#if $user.role.includes('manager')}
						<li><a href="/create-auction">Create Auction</a></li>
						<li><a href="/manage-users">Manage users</a></li>
						<li><a href="/auction-results">Auction results</a></li>
					{/if}
					<li><button type="button" on:click={logout} aria-label="Logout">Logout</button></li>
				</ul>
			</div>
		</div>
		{:else}
			<button class="btn btn-primary" on:click={loginWithDiscord}>Login with Discord</button>
		{/if}
	</div>
</nav>
{#if $openNotification}
<dialog class="modal modal-open">
  <div class="modal-box">
    <h3 class="text-lg font-bold pb-2">Notifications</h3>
    <NotificationsList bind:notificationCount bind:this={notificationListComponent} />
	    <div class="modal-action">
      <form method="dialog">  
		<button class="btn btn-error" on:click={()=>notificationListComponent.clearAll()}>Clear all</button>
		<!-- if there is a button in form, it will close the modal -->
        <button class="btn btn-primary" on:click={()=>{openNotification.set(false)}}>Close</button>
      </form>
    </div>
  </div>
</dialog>
{/if}
<ToastManager />
