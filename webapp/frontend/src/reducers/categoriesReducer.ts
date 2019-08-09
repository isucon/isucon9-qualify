import {AnyAction} from "redux";
import {CategorySimple} from "../dataObjects/category";
import {FETCH_SETTINGS_SUCCESS, FetchSettingsSuccessAction} from "../actions/settingsAction";

export interface CategoriesState {
    categories: CategorySimple[],
}

const initialState: CategoriesState = {
    categories: [],
};

type Actions = FetchSettingsSuccessAction | AnyAction;

const categories = (state: CategoriesState = initialState, action: Actions): CategoriesState => {
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