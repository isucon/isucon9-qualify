import { connect } from 'react-redux';
import { AppState } from '../index';
import BasePageComponent from '../components/BasePageComponent';
import { Dispatch } from 'redux';

const mapStateToProps = (state: AppState) => ({
  loading: state.page.isLoading,
  alreadyLoaded: state.authStatus.checked,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(BasePageComponent);
