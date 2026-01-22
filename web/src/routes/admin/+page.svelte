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

<div class="container mx-auto">
	<h1 class="mb-6 text-3xl font-bold">Admin Dashboard</h1>

	{#if loading}
		<div class="flex justify-center py-12">
			<span class="loading loading-spinner loading-lg"></span>
		</div>
	{:else if error}
		<div class="alert alert-error">
			<span>{error}</span>
		</div>
	{:else if stats}
		<!-- User Statistics -->
		<div class="mb-6">
			<h2 class="mb-3 text-2xl font-semibold">User Statistics</h2>
			<div class="stats stats-vertical shadow lg:stats-horizontal">
				<div class="stat">
					<div class="stat-title">Total Users</div>
					<div class="stat-value">{stats.totalUsers}</div>
				</div>
				<div class="stat">
					<div class="stat-title">Validated Users</div>
					<div class="stat-value">{stats.validatedUsers}</div>
					<div class="stat-desc">
						{((stats.validatedUsers / stats.totalUsers) * 100).toFixed(1)}% validated
					</div>
				</div>
			</div>
		</div>

		<!-- Token Statistics -->
		<div class="mb-6">
			<h2 class="mb-3 text-2xl font-semibold">Token Statistics</h2>
			<div class="stats stats-vertical shadow lg:stats-horizontal">
				<div class="stat">
					<div class="stat-title">Total Tokens</div>
					<div class="stat-value">{stats.totalTokens.toLocaleString()}</div>
				</div>
				<div class="stat">
					<div class="stat-title">Reserved Tokens</div>
					<div class="stat-value text-warning">{stats.totalReservedTokens.toLocaleString()}</div>
					<div class="stat-desc">In active bids</div>
				</div>
				<div class="stat">
					<div class="stat-title">Available Tokens</div>
					<div class="stat-value text-success">{stats.availableTokens.toLocaleString()}</div>
					<div class="stat-desc">
						{((stats.availableTokens / stats.totalTokens) * 100).toFixed(1)}% available
					</div>
				</div>
			</div>
		</div>

		<!-- Auction Statistics -->
		<div class="mb-6">
			<h2 class="mb-3 text-2xl font-semibold">Auction Statistics</h2>
			<div class="stats stats-vertical shadow lg:stats-horizontal">
				<div class="stat">
					<div class="stat-title">Total Auctions</div>
					<div class="stat-value">{stats.totalAuctions}</div>
				</div>
				<div class="stat">
					<div class="stat-title">Ongoing Auctions</div>
					<div class="stat-value text-info">{stats.ongoingAuctions}</div>
				</div>
				<div class="stat">
					<div class="stat-title">Finished Auctions</div>
					<div class="stat-value">{stats.finishedAuctions}</div>
				</div>
				<div class="stat">
					<div class="stat-title">Recent Auctions</div>
					<div class="stat-value text-primary">{stats.recentAuctionsCount}</div>
					<div class="stat-desc">Last 24 hours</div>
				</div>
			</div>
		</div>

		<!-- Bid and Result Statistics -->
		<div class="mb-6">
			<h2 class="mb-3 text-2xl font-semibold">Activity Statistics</h2>
			<div class="stats stats-vertical shadow lg:stats-horizontal">
				<div class="stat">
					<div class="stat-title">Total Bids</div>
					<div class="stat-value">{stats.totalBids}</div>
				</div>
				<div class="stat">
					<div class="stat-title">Unresolved Results</div>
					<div class="stat-value" class:text-warning={stats.unresolvedResults > 0}>
						{stats.unresolvedResults}
					</div>
					<div class="stat-desc">
						{#if stats.unresolvedResults > 0}
							<a href="/auction-results" class="link">Review results</a>
						{:else}
							All results resolved
						{/if}
					</div>
				</div>
				<div class="stat">
					<div class="stat-title">Unseen Notifications</div>
					<div class="stat-value">{stats.unseenNotifications}</div>
					<div class="stat-desc">of {stats.totalNotifications} total</div>
				</div>
			</div>
		</div>

		<!-- System Health -->
		<div class="mb-6">
			<h2 class="mb-3 text-2xl font-semibold">System Health</h2>
			<div class="stats shadow">
				<div class="stat">
					<div class="stat-title">Token Health Check</div>
					<div class="stat-value">
						<span
							class="badge badge-lg {getHealthCheckBadgeClass(stats.latestHealthCheckState)}"
						>
							{stats.latestHealthCheckState.toUpperCase()}
						</span>
					</div>
					<div class="stat-desc">Last check: {formatDate(stats.latestHealthCheckDate)}</div>
				</div>
			</div>
		</div>

		<!-- Quick Actions -->
		<div class="mb-6">
			<h2 class="mb-3 text-2xl font-semibold">Quick Actions</h2>
			<div class="flex flex-wrap gap-2">
				<a href="/create-auction" class="btn btn-primary">Create Auction</a>
				<a href="/manage-users" class="btn btn-secondary">Manage Users</a>
				<a href="/auction-results" class="btn btn-accent">Auction Results</a>
				<button class="btn btn-info" on:click={fetchDashboardStats}>Refresh Stats</button>
			</div>
		</div>
	{/if}
</div>

<AuthGuard requiredRole="manager" />
