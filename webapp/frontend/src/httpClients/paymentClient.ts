import config from '../config';

/**
 * HTTP client for payment service
 */
class PaymentClient {
  private baseUrl: string = config.paymentUrl;
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
      mode: 'cors',
      headers: Object.assign({}, this.defaultHeaders, {
        'Content-Type': 'application/json',
      }),
      credentials: 'same-origin',
    };

    if (params) {
      requestOption.body = JSON.stringify(params);
    }

    return await fetch(`${this.baseUrl}${path}`, requestOption);
  }
}

export default new PaymentClient();
