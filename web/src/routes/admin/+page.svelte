<script lang="ts">
	import { onMount } from 'svelte';
	import pb from '$lib/pocketbase';
	import AuthGuard from '../../components/AuthGuard.svelte';

	interface DashboardStats {
		totalUsers: number;
		validatedUsers: number;
		totalTokens: number;
		totalReservedTokens: number;
		availableTokens: number;
		ongoingAuctions: number;
		finishedAuctions: number;
		totalAuctions: number;
		recentAuctionsCount: number;
		totalBids: number;
		unresolvedResults: number;
		latestHealthCheckState: string;
		latestHealthCheckDate: string | null;
		totalNotifications: number;
		unseenNotifications: number;
	}

	let stats: DashboardStats | null = null;
	let loading = true;
	let error = '';

	async function fetchDashboardStats() {
		loading = true;
		error = '';
		try {
			const response = await pb.send('/api/dashboard-stats', {
				method: 'GET'
			});
			stats = response as DashboardStats;
		} catch (err) {
			console.error('Error fetching dashboard stats:', err);
			error = 'Failed to load dashboard statistics';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		fetchDashboardStats();
		// Refresh stats every 30 seconds
		const interval = setInterval(fetchDashboardStats, 30000);
		return () => clearInterval(interval);
	});

	function getHealthCheckBadgeClass(state: string): string {
		switch (state) {
			case 'ok':
				return 'badge-success';
			case 'error':
				return 'badge-error';
			default:
				return 'badge-warning';
		}
	}

	function formatDate(dateString: string | null): string {
		if (!dateString) return 'N/A';
		const date = new Date(dateString);
		return date.toLocaleString();
	}
</script>

<div class="space-y-6">
	<!-- Page Header -->
	<div class="bg-gradient-to-r from-primary/20 to-secondary/20 rounded-box p-6 shadow-lg border border-base-content/10">
		<h1 class="text-2xl font-bold mb-2 flex items-center gap-2">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
			</svg>
			Admin Dashboard
		</h1>
		<p class="text-sm opacity-80">Monitor system metrics and manage auctions</p>
	</div>

	{#if loading}
		<div class="flex justify-center py-16 bg-base-200 rounded-box">
			<div class="flex flex-col items-center gap-4">
				<span class="loading loading-spinner loading-lg text-primary"></span>
				<p class="text-lg opacity-70">Loading dashboard statistics...</p>
			</div>
		</div>
	{:else if error}
		<div class="alert alert-error shadow-lg">
			<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
				<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" />
			</svg>
			<span>{error}</span>
		</div>
	{:else if stats}
		<!-- User Statistics -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h2 class="text-2xl font-semibold mb-4 flex items-center gap-3">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
				</svg>
				User Statistics
			</h2>
			<div class="stats stats-vertical shadow-xl lg:stats-horizontal w-full border border-base-content/10">
				<div class="stat bg-primary/10">
					<div class="stat-figure text-primary">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
					</div>
					<div class="stat-title">Total Users</div>
					<div class="stat-value text-primary">{stats.totalUsers}</div>
				</div>
				<div class="stat bg-success/10">
					<div class="stat-figure text-success">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div class="stat-title">Validated Users</div>
					<div class="stat-value text-success">{stats.validatedUsers}</div>
					<div class="stat-desc font-semibold">
						{stats.totalUsers > 0 ? ((stats.validatedUsers / stats.totalUsers) * 100).toFixed(1) : 0}% validated
					</div>
				</div>
			</div>
		</div>

		<!-- Token Statistics -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h2 class="text-2xl font-semibold mb-4 flex items-center gap-3">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
				</svg>
				Token Statistics
			</h2>
			<div class="stats stats-vertical shadow-xl lg:stats-horizontal w-full border border-base-content/10">
				<div class="stat bg-info/10">
					<div class="stat-figure text-info">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 9V7a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2m2 4h10a2 2 0 002-2v-6a2 2 0 00-2-2H9a2 2 0 00-2 2v6a2 2 0 002 2zm7-5a2 2 0 11-4 0 2 2 0 014 0z" />
						</svg>
					</div>
					<div class="stat-title">Total Tokens</div>
					<div class="stat-value text-info">{stats.totalTokens.toLocaleString()}</div>
				</div>
				<div class="stat bg-warning/10">
					<div class="stat-figure text-warning">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
						</svg>
					</div>
					<div class="stat-title">Reserved Tokens</div>
					<div class="stat-value text-warning">{stats.totalReservedTokens.toLocaleString()}</div>
					<div class="stat-desc font-semibold">In active bids</div>
				</div>
				<div class="stat bg-success/10">
					<div class="stat-figure text-success">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div class="stat-title">Available Tokens</div>
					<div class="stat-value text-success">{stats.availableTokens.toLocaleString()}</div>
					<div class="stat-desc font-semibold">
						{stats.totalTokens > 0 ? ((stats.availableTokens / stats.totalTokens) * 100).toFixed(1) : 0}% available
					</div>
				</div>
			</div>
		</div>

		<!-- Auction Statistics -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h2 class="text-2xl font-semibold mb-4 flex items-center gap-3">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
				</svg>
				Auction Statistics
			</h2>
			<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
				<div class="stat bg-base-100 shadow-xl rounded-box border border-base-content/10">
					<div class="stat-figure text-accent">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
						</svg>
					</div>
					<div class="stat-title">Total Auctions</div>
					<div class="stat-value text-accent">{stats.totalAuctions}</div>
				</div>
				<div class="stat bg-base-100 shadow-xl rounded-box border border-info/30">
					<div class="stat-figure text-info">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div class="stat-title">Ongoing Auctions</div>
					<div class="stat-value text-info">{stats.ongoingAuctions}</div>
				</div>
				<div class="stat bg-base-100 shadow-xl rounded-box border border-base-content/10">
					<div class="stat-figure opacity-50">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div class="stat-title">Finished Auctions</div>
					<div class="stat-value">{stats.finishedAuctions}</div>
				</div>
				<div class="stat bg-base-100 shadow-xl rounded-box border border-primary/30">
					<div class="stat-figure text-primary">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
						</svg>
					</div>
					<div class="stat-title">Recent Auctions</div>
					<div class="stat-value text-primary">{stats.recentAuctionsCount}</div>
					<div class="stat-desc font-semibold">Last 24 hours</div>
				</div>
			</div>
		</div>

		<!-- Bid and Result Statistics -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h2 class="text-2xl font-semibold mb-4 flex items-center gap-3">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z" />
				</svg>
				Activity Statistics
			</h2>
			<div class="stats stats-vertical shadow-xl lg:stats-horizontal w-full border border-base-content/10">
				<div class="stat bg-base-100">
					<div class="stat-figure text-accent">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4" />
						</svg>
					</div>
					<div class="stat-title">Total Bids</div>
					<div class="stat-value">{stats.totalBids}</div>
				</div>
				<div class="stat bg-base-100">
					<div class="stat-figure" class:text-warning={stats.unresolvedResults > 0}>
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div class="stat-title">Unresolved Results</div>
					<div class="stat-value" class:text-warning={stats.unresolvedResults > 0}>
						{stats.unresolvedResults}
					</div>
					<div class="stat-desc">
						{#if stats.unresolvedResults > 0}
							<a href="/auction-results" class="link link-warning font-semibold">Review results →</a>
						{:else}
							<span class="text-success font-semibold">✓ All results resolved</span>
						{/if}
					</div>
				</div>
				<div class="stat bg-base-100">
					<div class="stat-figure text-info">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9" />
						</svg>
					</div>
					<div class="stat-title">Unseen Notifications</div>
					<div class="stat-value text-info">{stats.unseenNotifications}</div>
					<div class="stat-desc font-semibold">of {stats.totalNotifications} total</div>
				</div>
			</div>
		</div>

		<!-- System Health -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h2 class="text-2xl font-semibold mb-4 flex items-center gap-3">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
				</svg>
				System Health
			</h2>
			<div class="stats shadow-xl w-full border border-base-content/10">
				<div class="stat">
					<div class="stat-figure">
						<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor" class:text-success={stats.latestHealthCheckState === 'ok'} class:text-error={stats.latestHealthCheckState === 'error'} class:text-warning={stats.latestHealthCheckState !== 'ok' && stats.latestHealthCheckState !== 'error'}>
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
						</svg>
					</div>
					<div class="stat-title">Token Health Check</div>
					<div class="stat-value">
						<span
							class="badge badge-lg {getHealthCheckBadgeClass(stats.latestHealthCheckState)}"
						>
							{stats.latestHealthCheckState.toUpperCase()}
						</span>
					</div>
					<div class="stat-desc font-semibold">Last check: {formatDate(stats.latestHealthCheckDate)}</div>
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="bg-base-200 rounded-box p-6 shadow-lg border border-base-content/10">
			<h2 class="text-2xl font-semibold mb-4 flex items-center gap-3">
				<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z" />
				</svg>
				Quick Actions
			</h2>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
				<a href="/create-auction" class="btn btn-primary btn-lg gap-2 shadow-lg">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
					</svg>
					Create Auction
				</a>
				<a href="/manage-users" class="btn btn-secondary btn-lg gap-2 shadow-lg">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
					</svg>
					Manage Users
				</a>
				<a href="/auction-results" class="btn btn-accent btn-lg gap-2 shadow-lg">
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
					</svg>
					Auction Results
				</a>
				<button class="btn btn-info btn-lg gap-2 shadow-lg" on:click={fetchDashboardStats}>
					<svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
					</svg>
					Refresh Stats
				</button>
			</div>
		</div>
	{/if}
</div>

<AuthGuard requiredRole="admin" />
