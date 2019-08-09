import { connect } from 'react-redux';
import { AppState } from '../index';
import ItemListPage from '../pages/ItemListPage';

const mapStateToProps = (state: AppState) => ({
  items: state.timeline.items,
  errorType: state.error.errorType,
  loading: false, // TODO state.page.isLoading,
});
const mapDispatchToProps = (dispatch: any) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(ItemListPage);
