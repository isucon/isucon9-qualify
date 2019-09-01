import { connect } from 'react-redux';
import { AppState } from '../index';
import { Dispatch } from 'redux';
import { closeSnackBarAction } from '../actions/snackBarAction';
import { SnackBar } from '../components/SnackBar';
import * as React from 'react';

const mapStateToProps = (state: AppState) => ({
  open: state.snackBar.available,
  message: state.snackBar.reason,
  variant: state.snackBar.variant,
});
const mapDispatchToProps = (dispatch: Dispatch) => ({
  handleClose(event: React.MouseEvent) {
    dispatch(closeSnackBarAction());
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SnackBar);
