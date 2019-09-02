import { connect } from 'react-redux';
import { AppState } from '../index';
import { NotFoundPage } from '../pages/error/NotFoundPage';
import { Dispatch } from 'redux';

const mapStateToProps = (state: AppState) => ({
  message: state.error.errorMessage,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(NotFoundPage);
