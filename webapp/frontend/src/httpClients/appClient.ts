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

    async post(path: string, params?: Object): Promise<Response> {
        let requestOption: RequestInit = {
            method: 'POST',
            mode: 'same-origin',
            headers: Object.assign({}, this.defaultHeaders, {
                'Content-Type': 'application/json',
            }),
        };

        if (params) {
            const body = JSON.stringify(params);
            requestOption.body = body;
        }


        return await fetch(`${this.baseUrl}${path}`, requestOption);
    }
}

export default new AppClient();