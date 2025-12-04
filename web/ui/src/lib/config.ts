interface AppConfig {
    apiURL: string;
}

declare global {
    interface Window {
        __APP_CONFIG__?: AppConfig;
    }
}

export const config: AppConfig = {
    apiURL: window.__APP_CONFIG__?.apiURL || 'http://localhost:3000',
};
