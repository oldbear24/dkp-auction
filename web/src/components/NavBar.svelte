<script lang="ts">
	import { user } from '$lib/stores/store';
	import { onMount } from 'svelte';
	import pb, { subscribeToUserUpdate, unsubscribeFromUserUpdates } from '$lib/pocketbase';
	import ToastManager from './ToastManager.svelte';
	import { writable } from 'svelte/store';
	import NotificationsList from './NotificationsList.svelte';
	import { appVersion } from '$lib/version';
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

<nav class="navbar bg-base-300 shadow-lg border-b border-base-content/10">
	<div class="flex-1">
		<a class="btn btn-ghost text-xl font-bold hover:bg-base-content/10 transition-colors" href="/">
			🏛️ Auction House
		</a>
	</div>

	<div class="flex-none pr-4 flex items-center gap-3">
		<span class="text-xs opacity-60 hidden sm:inline">{appVersion.version}</span>
		{#if $user}
			<div
				data-tip="Your current tokens and usable tokens."
				class="tooltip tooltip-left tooltip-info"
			>
				<div class="badge badge-lg badge-primary gap-2 font-semibold">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					{$user.tokens - $user.reservedTokens} / {$user.tokens}
				</div>
			</div>
			<div class="indicator">
				{#if $notificationCount > 99}
					<span class="indicator-item badge badge-secondary badge-sm">99+</span>
				{:else if $notificationCount>0}
					<span class="indicator-item badge badge-secondary badge-sm animate-pulse">{$notificationCount}</span>
				{/if}
			<div class="dropdown dropdown-end">
				<button class="btn btn-ghost btn-circle avatar avatar-placeholder hover:ring-2 hover:ring-primary transition-all" aria-label="User menu">
					<div class="ring-primary ring-offset-base-100 w-11 rounded-full ring-2 ring-offset-2">
						{#if $user.avatar}
							<img
								src={pb.files.getURL($user, $user.avatar, { thumb: '100x100' })}
								alt="User Avatar"
							/>
						{:else}
							<span class="text-xl font-bold">{$user.name.substring(0, 2)}</span>
						{/if}
					</div>
				</button>
				<ul
					class="menu dropdown-content bg-base-200 rounded-box mt-3 w-56 p-2 shadow-xl border border-base-content/10"
					style="z-index: 100;"
				>
					<li><a href="/profile" class="gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
						</svg>
						Profile
					</a></li>
          <li><button on:click="{() => {openNotification.set(true)}}" class="gap-2">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
            Notifications 
						<div class="badge badge-secondary badge-sm">{$notificationCount}</div>
          </button></li>
		  			{#if $user.role.includes('admin')}
						<div class="divider my-1"></div>
						<li class="menu-title"><span class="text-xs font-semibold">Admin</span></li>
		  				<li><a href="/admin" class="gap-2">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
							</svg>
							Dashboard
						</a></li>
					{/if}

					{#if $user.role.includes('manager')}
						<div class="divider my-1"></div>
						<li class="menu-title"><span class="text-xs font-semibold">Manager</span></li>
						<li><a href="/create-auction" class="gap-2">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
							</svg>
							Create Auction
						</a></li>
						<li><a href="/manage-users" class="gap-2">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
							</svg>
							Manage Users
						</a></li>
						<li><a href="/auction-results" class="gap-2">
							<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
								<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
							</svg>
							Auction Results
						</a></li>
					{/if}
					<div class="divider my-1"></div>
					<li><button type="button" on:click={logout} aria-label="Logout" class="gap-2 text-error">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
						</svg>
						Logout
					</button></li>
				</ul>
			</div>
		</div>
		{:else}
			<button class="btn btn-primary gap-2 shadow-lg" on:click={loginWithDiscord}>
				<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
					<path d="M20.317 4.37a19.791 19.791 0 00-4.885-1.515.074.074 0 00-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 00-5.487 0 12.64 12.64 0 00-.617-1.25.077.077 0 00-.079-.037A19.736 19.736 0 003.677 4.37a.07.07 0 00-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 00.031.057 19.9 19.9 0 005.993 3.03.078.078 0 00.084-.028 14.09 14.09 0 001.226-1.994.076.076 0 00-.041-.106 13.107 13.107 0 01-1.872-.892.077.077 0 01-.008-.128 10.2 10.2 0 00.372-.292.074.074 0 01.077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 01.078.01c.12.098.246.198.373.292a.077.077 0 01-.006.127 12.299 12.299 0 01-1.873.892.077.077 0 00-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 00.084.028 19.839 19.839 0 006.002-3.03.077.077 0 00.032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 00-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.956-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419 0-1.333.955-2.419 2.157-2.419 1.21 0 2.176 1.096 2.157 2.42 0 1.333-.946 2.418-2.157 2.418z"/>
				</svg>
				Login with Discord
			</button>
		{/if}
	</div>
</nav>
{#if $openNotification}
<dialog class="modal modal-open">
  <div class="modal-box max-w-2xl shadow-2xl">
    <h3 class="text-2xl font-bold pb-4 flex items-center gap-2">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
			</svg>
			Notifications
		</h3>
    <NotificationsList bind:notificationCount bind:this={notificationListComponent} />
	    <div class="modal-action gap-2">
      <form method="dialog" class="flex gap-2">  
		<button class="btn btn-error gap-2" on:click={()=>notificationListComponent.clearAll()}>
			<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
			</svg>
			Clear All
		</button>
        <button class="btn btn-primary gap-2" on:click={()=>{openNotification.set(false)}}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
					</svg>
					Close
				</button>
      </form>
    </div>
  </div>
</dialog>
{/if}
<ToastManager />
