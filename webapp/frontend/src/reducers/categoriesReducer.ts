import { CategorySimple } from '../dataObjects/category';
import { FETCH_SETTINGS_SUCCESS } from '../actions/settingsAction';
import { ActionTypes } from '../actions/actionTypes';

export interface CategoriesState {
  categories: CategorySimple[];
}

const initialState: CategoriesState = {
  categories: [],
};

const categories = (
  state: CategoriesState = initialState,
  action: ActionTypes,
): CategoriesState => {
  switch (action.type) {
    case FETCH_SETTINGS_SUCCESS:
      return {
        categories: action.payload.settings.categories,
      };
    default:
      return { ...state };
  }
};

export default categories;
