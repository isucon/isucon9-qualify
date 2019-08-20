import React, { PropsWithChildren } from 'react';

import {
  Container,
  MuiThemeProvider,
  createMuiTheme,
  Theme,
  WithStyles,
} from '@material-ui/core';
import LoadingComponent from './LoadingComponent';
import HeaderContainer from '../containers/HeaderContainer';
import { StyleRules } from '@material-ui/core/styles';
import createStyles from '@material-ui/core/styles/createStyles';
import withStyles from '@material-ui/core/styles/withStyles';

const themeInstance = createMuiTheme({
  palette: {
    background: {
      default: 'white',
    },
  },
});

const styles = (theme: Theme): StyleRules =>
  createStyles({
    container: {
      paddingTop: theme.spacing(7),
    },
  });

interface BaseProps extends WithStyles<typeof styles> {
  load: () => void;
  alreadyLoaded: boolean;
  loading: boolean;
}

export type Props = PropsWithChildren<BaseProps>;

class BasePageComponent extends React.Component<Props> {
  constructor(props: Props) {
    super(props);

    if (!this.props.alreadyLoaded) {
      this.props.load();
    }
  }

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
        </Container>
      </MuiThemeProvider>
    );
  }
}

export default withStyles(styles)(BasePageComponent);
