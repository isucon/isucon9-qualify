export class ResponseError extends Error {
  private readonly res: Response | undefined;

  constructor(message: string, response?: Response) {
    super(message);
    this.res = response;
  }

  getResponse(): Response | undefined {
    return this.res;
  }
}
