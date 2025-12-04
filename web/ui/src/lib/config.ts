interface AppConfig {
    apiURL: string;
}

declare global {
    interface Window {
        __APP_CONFIG__?: AppConfig;
    }
}

export const config: AppConfig = {
    apiURL: window.__APP_CONFIG__?.apiURL || import.meta.env.VITE_API_URL || 'http://localhost:8080',
};
