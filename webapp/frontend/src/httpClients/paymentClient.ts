/**
 * HTTP client for payment service
 */
class PaymentClient {
  private baseURL?: string;
  private defaultHeaders: HeadersInit = {};

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

    return await fetch(`${this.baseURL}${path}`, requestOption);
  }

  public setBaseURL(baseURL: string) {
    this.baseURL = baseURL;
  }
}

export default new PaymentClient();
