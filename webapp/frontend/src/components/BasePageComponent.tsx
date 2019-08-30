import React, { PropsWithChildren } from 'react';
import {
  Container,
  MuiThemeProvider,
  Theme,
  WithStyles,
} from '@material-ui/core';
import LoadingComponent from './LoadingComponent';
import HeaderContainer from '../containers/HeaderContainer';
import SnackBarContainer from '../containers/SnackBarContainer';
import { StyleRules } from '@material-ui/core/styles';
import createStyles from '@material-ui/core/styles/createStyles';
import withStyles from '@material-ui/core/styles/withStyles';
import { themeInstance } from '../theme';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    container: {
      paddingTop: theme.spacing(12),
    },
  });

interface BaseProps extends WithStyles<typeof styles> {
  loading: boolean;
}

export type Props = PropsWithChildren<BaseProps>;

class BasePageComponent extends React.Component<Props> {
  render() {
    const { classes } = this.props;

    return (
      <MuiThemeProvider theme={themeInstance}>
        <Container maxWidth="md" className={classes.container}>
          <HeaderContainer />
          {this.props.loading ? (
            <LoadingComponent />
          ) : (
            this.props.children || null
          )}
          <SnackBarContainer />
        </Container>
      </MuiThemeProvider>
    );
  }
}

export default withStyles(styles)(BasePageComponent);
