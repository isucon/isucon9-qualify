import config from '../config';

/**
 * HTTP client for main app
 */
class AppClient {
    private baseUrl: string = config.apiUrl;
    private defaultHeaders: HeadersInit = {};

    async get(path: string): Promise<Response> {
        return await fetch(`${this.baseUrl}${path}`, {
            method: 'GET',
            headers: this.defaultHeaders,
        });
    }

    async post(path: string): Promise<Response> {
        return await fetch(`${this.baseUrl}${path}`, {
            method: 'POST',
            mode: 'cors',
            headers: this.defaultHeaders,
        });
    }
}

export default new AppClient();