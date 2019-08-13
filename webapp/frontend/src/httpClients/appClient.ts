import config from '../config';
import { SettingsRes } from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';

/**
 * HTTP client for main app
 */
class AppClient {
  private baseUrl: string = config.apiUrl;
  private defaultHeaders: HeadersInit = {};

  async get(path: string, params: Record<string, any> = {}): Promise<Response> {
    let getParams = new URLSearchParams();
    for (const key in params) {
      const value = params[key];
      if (value !== undefined) {
        getParams.set(key, params[key]);
      }
    }

    let url = `${this.baseUrl}${path}`;
    if (!(Object.keys(params).length === 0)) {
      url = `${url}?${getParams.toString()}`;
    }

    return await fetch(url, {
      method: 'GET',
      headers: this.defaultHeaders,
    });
  }

  async post(path: string, params: any = {}): Promise<Response> {
    let requestOption: RequestInit = {
      method: 'POST',
      mode: 'same-origin',
      headers: Object.assign({}, this.defaultHeaders, {
        'Content-Type': 'application/json',
      }),
      credentials: 'same-origin',
    };

    params.csrf_token = await this.getCsrfToken();

    if (params) {
      requestOption.body = JSON.stringify(params);
    }

    return await fetch(`${this.baseUrl}${path}`, requestOption);
  }

  private async getCsrfToken(): Promise<string> {
    const res: Response = await fetch('/settings', {
      method: 'GET',
      headers: this.defaultHeaders,
    });

    if (!res.ok) {
      throw new AppResponseError('CSRF tokenの取得に失敗しました', res);
    }

    const body: SettingsRes = await res.json();

    return body.csrf_token;
  }
}

export default new AppClient();
