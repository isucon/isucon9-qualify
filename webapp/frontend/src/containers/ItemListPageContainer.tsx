import { connect } from 'react-redux';
import { AppState } from '../index';
import ItemListPage from '../pages/ItemListPage';

const mapStateToProps = (state: AppState) => ({
  items: state.timeline.items,
  errorType: state.error.errorType,
  loading: state.page.isTimelineLoading,
});
const mapDispatchToProps = (dispatch: any) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ItemListPage);
