import { SettingsRes } from '../types/appApiTypes';
import { AppResponseError } from '../errors/AppResponseError';

/**
 * HTTP client for main app
 */
class AppClient {
  private defaultHeaders: HeadersInit = {};

  async get(path: string, params: Record<string, any> = {}): Promise<Response> {
    let getParams = new URLSearchParams();
    for (const key in params) {
      const value = params[key];
      if (value !== undefined) {
        getParams.set(key, params[key]);
      }
    }

    let url = `${path}`;
    if (getParams.toString() !== '') {
      url = `${url}?${getParams.toString()}`;
    }

    return await fetch(url, {
      method: 'GET',
      headers: this.defaultHeaders,
    });
  }

  async post(
    path: string,
    params: any = {},
    csrfCheckRequired: boolean = true,
  ): Promise<Response> {
    let requestOption: RequestInit = {
      method: 'POST',
      mode: 'same-origin',
      headers: Object.assign({}, this.defaultHeaders, {
        'Content-Type': 'application/json',
      }),
      credentials: 'same-origin',
    };

    if (csrfCheckRequired) {
      params.csrf_token = await this.getCsrfToken();
    }

    requestOption.body = JSON.stringify(params);

    return await fetch(path, requestOption);
  }

  async postFormData(path: string, body: FormData): Promise<Response> {
    let requestOption: RequestInit = {
      method: 'POST',
      mode: 'same-origin',
      // MEMO: The reason why we should not set Content-Type header by ourselves
      // https://stackoverflow.com/questions/39280438/fetch-missing-boundary-in-multipart-form-data-post
      headers: this.defaultHeaders,
      credentials: 'same-origin',
    };

    body.append('csrf_token', await this.getCsrfToken());
    requestOption.body = body;

    return await fetch(path, requestOption);
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
