export const appVersion = {
	version: import.meta.env.VITE_APP_VERSION ?? 'dev',
	commit: import.meta.env.VITE_APP_COMMIT ?? 'unknown',
	date: import.meta.env.VITE_APP_BUILD_DATE ?? 'unknown'
};
