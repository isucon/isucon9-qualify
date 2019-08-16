import { connect } from 'react-redux';
import SellFormComponent from '../components/SellFormComponent';
import { listItemAction } from '../actions/sellingItemAction';
import { AppState } from '../index';
import { AnyAction } from 'redux';
import { ThunkDispatch } from 'redux-thunk';
import { CategorySimple } from '../dataObjects/category';

const mapStateToProps = (state: AppState) => {
  // Note: Parent category's parent_id is 0
  const categories = state.categories.categories.filter(
    (category: CategorySimple) => category.parentId !== 0,
  );

  return {
    error: state.formError.error,
    categories,
  };
};
const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  sellItem: (
    name: string,
    description: string,
    price: number,
    categoryId: number,
    image: Blob,
  ) => {
    dispatch(listItemAction(name, description, price, categoryId, image));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SellFormComponent);
