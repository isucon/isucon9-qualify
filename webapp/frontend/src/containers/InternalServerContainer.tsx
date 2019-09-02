import { connect } from 'react-redux';
import { AppState } from '../index';
import { Dispatch } from 'redux';
import { InternalServerErrorPage } from '../pages/error/InternalServerErrorPage';

const mapStateToProps = (state: AppState) => ({
  message: state.error.errorMessage,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(InternalServerErrorPage);
