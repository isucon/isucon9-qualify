import React from 'react';

import {
  Avatar,
  Typography,
  TextField,
  Button,
  Grid,
  createStyles,
  WithStyles,
} from '@material-ui/core';
import { Theme } from '@material-ui/core/styles/createMuiTheme';
import { LockOutlined } from '@material-ui/icons';
import { Link as RouteLink } from 'react-router-dom';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import { routes } from '../routes/Route';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    paper: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
    },
    avatar: {
      margin: theme.spacing(1),
      backgroundColor: theme.palette.secondary.main,
    },
    form: {
      width: '100%',
      marginTop: theme.spacing(1),
    },
    submit: {
      margin: theme.spacing(3, 0, 2),
    },
  });

interface SignInFormComponentProps extends WithStyles<typeof styles> {
  onSubmit: (accountName: string, password: string) => void;
}

interface SignInFormComponentState {
  accountName: string;
  password: string;
}

class SignInPageFormComponent extends React.Component<
  SignInFormComponentProps,
  SignInFormComponentState
> {
  constructor(props: SignInFormComponentProps) {
    super(props);

    this.state = {
      accountName: '',
      password: '',
    };

    this._onSubmit = this._onSubmit.bind(this);
    this._onChangeAccountName = this._onChangeAccountName.bind(this);
    this._onChangePassword = this._onChangePassword.bind(this);
  }

  _onSubmit(e: React.MouseEvent) {
    e.preventDefault();
    const { accountName, password } = this.state;
    this.props.onSubmit(accountName, password);
  }

  _onChangeAccountName(e: React.ChangeEvent<HTMLInputElement>) {
    this.setState({
      accountName: e.target.value,
    });
  }

  _onChangePassword(e: React.ChangeEvent<HTMLInputElement>) {
    this.setState({
      password: e.target.value,
    });
  }

  render() {
    const { accountName, password } = this.state;
    const { classes } = this.props;

    return (
      <div className={classes.paper}>
        <Avatar className={classes.avatar}>
          <LockOutlined />
        </Avatar>
        <Typography component="h1" variant="h5">
          ログインページ
        </Typography>
        <form className={classes.form} noValidate>
          <TextField
            variant="outlined"
            margin="normal"
            required
            fullWidth
            id="accountName"
            label="ユーザ名"
            name="accountName"
            autoFocus
            value={accountName}
            onChange={this._onChangeAccountName}
          />
          <TextField
            variant="outlined"
            margin="normal"
            required
            fullWidth
            id="password"
            label="パスワード"
            name="password"
            type="password"
            autoComplete="current-password"
            value={password}
            onChange={this._onChangePassword}
          />
          <Button
            id="signInButton"
            type="submit"
            fullWidth
            variant="contained"
            color="primary"
            onClick={this._onSubmit}
            className={classes.submit}
          >
            ログイン
          </Button>
          <Grid container>
            <Grid item xs>
              <RouteLink to={routes.register.path}>新規登録はこちら</RouteLink>
            </Grid>
          </Grid>
        </form>
      </div>
    );
  }
}

export default withStyles(styles)(SignInPageFormComponent);
