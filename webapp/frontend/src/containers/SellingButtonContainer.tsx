import { push } from 'connected-react-router';
import { SellingButtonComponent } from '../components/SellingButtonComponent';
import { connect } from 'react-redux';
import { routes } from '../routes/Route';
import * as React from 'react';
import { AppState } from '../index';
import { ThunkDispatch } from 'redux-thunk';
import { AnyAction } from 'redux';

const mapStateToProps = (state: AppState) => ({});

const mapDispatchToProps = (
  dispatch: ThunkDispatch<AppState, undefined, AnyAction>,
) => ({
  onClick: (e: React.MouseEvent) => {
    e.preventDefault();
    dispatch(push(routes.sell.path));
  },
});

export default connect(
  mapStateToProps,
  mapDispatchToProps,
)(SellingButtonComponent);
