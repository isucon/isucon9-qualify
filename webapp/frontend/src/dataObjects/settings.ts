import { UserData } from './user';
import { CategorySimple } from './category';

export interface Settings {
  csrfToken: string;
  categories: CategorySimple[];
  user?: UserData;
}
