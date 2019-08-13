import { connect } from 'react-redux';
import { AppState } from '../index';
import ItemListPage from '../pages/ItemListPage';
import { fetchTimelineAction } from '../actions/fetchTimelineAction';

const mapStateToProps = (state: AppState) => {
  return {
    items: state.timeline.items,
    hasNext: state.timeline.hasNext,
    errorType: state.error.errorType,
    loading: state.page.isTimelineLoading,
  };
};
const mapDispatchToProps = (dispatch: any) => ({
  load: () => {
    dispatch(fetchTimelineAction());
  },
  loadMore: (createdAt: number, itemId: number, page: number) => {
    dispatch(fetchTimelineAction(createdAt, itemId));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ItemListPage);
