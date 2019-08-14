import { connect } from 'react-redux';
import { AppState } from '../index';
import { fetchTimelineAction } from '../actions/fetchTimelineAction';
import CategoryItemListPage from '../pages/CategoryItemListPage';

const mapStateToProps = (state: AppState) => {
  return {
    items: state.timeline.items,
    hasNext: state.timeline.hasNext,
    categoryId: state.timeline.categoryId,
    categoryName: state.timeline.categoryName,
    errorType: state.error.errorType,
    loading: state.page.isTimelineLoading,
  };
};
const mapDispatchToProps = (dispatch: any) => ({
  load: (categoryId: number) => {
    dispatch(fetchTimelineAction(undefined, undefined, categoryId));
  },
  loadMore: (
    createdAt: number,
    itemId: number,
    categoryId: number,
    page: number,
  ) => {
    dispatch(fetchTimelineAction(createdAt, itemId, categoryId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(CategoryItemListPage);
