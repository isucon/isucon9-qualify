import { ErrorRes } from '../types/appApiTypes';
import { NotFoundError } from '../errors/NotFoundError';
import { InternalServerError } from '../errors/InternalServerError';
import { AppResponseError } from '../errors/AppResponseError';

/**
 * checking response from application and throw error if it's necessary
 */
export async function checkAppResponse(response: Response) {
  if (!response.ok) {
    const errRes: ErrorRes = await response.json();

    if (response.status === 404) {
      throw new NotFoundError(errRes.error);
    }

    if (response.status >= 500) {
      throw new InternalServerError(errRes.error);
    }

    throw new AppResponseError(errRes.error, response);
  }
}
