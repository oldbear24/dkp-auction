<script lang="ts">
	import { goto } from '$app/navigation';
	import pb from '$lib/pocketbase';

	export let requiredRole: string = '';
	let userRoles: string[] = pb.authStore.record?.role ?? [];
    console.debug("User roles:",userRoles,"Required role",requiredRole)
	if (!pb.authStore.record || !pb.authStore.isValid || (requiredRole != '' && !userRoles.includes(requiredRole))) {
		goto('/');
	}
</script>
