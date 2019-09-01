import React from 'react';
import {
  Avatar,
  createStyles,
  Theme,
  Typography,
  WithStyles,
} from '@material-ui/core';
import { LockOutlined } from '@material-ui/icons';
import TextField from '@material-ui/core/TextField';
import Button from '@material-ui/core/Button';
import Grid from '@material-ui/core/Grid';
import { Link as RouteLink } from 'react-router-dom';
import { StyleRules } from '@material-ui/core/styles';
import withStyles from '@material-ui/core/styles/withStyles';
import { RegisterReq } from '../types/appApiTypes';
import { routes } from '../routes/Route';

const styles = (theme: Theme): StyleRules =>
  createStyles({
    paper: {
      marginTop: theme.spacing(1),
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

interface SignUpFormComponentProps extends WithStyles<typeof styles> {
  register: (params: RegisterReq) => void;
}

interface SignUpFormComponentState {
  accountName: string;
  address: string;
  password: string;
}

class SignUpFormComponent extends React.Component<
  SignUpFormComponentProps,
  SignUpFormComponentState
> {
  constructor(props: SignUpFormComponentProps) {
    super(props);

    this.state = {
      accountName: '',
      address: '',
      password: '',
    };

    this._onSubmit = this._onSubmit.bind(this);
    this._onChangeAccountName = this._onChangeAccountName.bind(this);
    this._onChangeAddress = this._onChangeAddress.bind(this);
    this._onChangePassword = this._onChangePassword.bind(this);
  }

  _onSubmit(e: React.MouseEvent) {
    e.preventDefault();
    this.props.register({
      account_name: this.state.accountName,
      address: this.state.address,
      password: this.state.password,
    });
  }

  _onChangeAccountName(e: React.ChangeEvent<HTMLInputElement>) {
    this.setState({
      accountName: e.target.value,
    });
  }

  _onChangeAddress(e: React.ChangeEvent<HTMLInputElement>) {
    this.setState({
      address: e.target.value,
    });
  }

  _onChangePassword(e: React.ChangeEvent<HTMLInputElement>) {
    this.setState({
      password: e.target.value,
    });
  }

  render() {
    const { classes } = this.props;
    const { accountName, address, password } = this.state;

    return (
      <div className={classes.paper}>
        <Avatar className={classes.avatar}>
          <LockOutlined />
        </Avatar>
        <Typography component="h1" variant="h5">
          新規登録
        </Typography>
        <form className={classes.form} noValidate>
          <TextField
            variant="outlined"
            margin="normal"
            required
            fullWidth
            id="name"
            label="ユーザ名"
            name="name"
            value={accountName}
            onChange={this._onChangeAccountName}
            autoFocus
          />
          <TextField
            variant="outlined"
            margin="normal"
            required
            fullWidth
            id="address"
            label="住所"
            name="address"
            value={address}
            onChange={this._onChangeAddress}
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
            value={password}
            onChange={this._onChangePassword}
          />
          <Button
            id="signUpButton"
            type="submit"
            fullWidth
            variant="contained"
            color="primary"
            className={classes.submit}
            onClick={this._onSubmit}
          >
            新規登録
          </Button>
          <Grid container>
            <Grid item xs>
              <RouteLink to={routes.login.path}>
                すでにアカウントをお持ちの方はこちら
              </RouteLink>
            </Grid>
          </Grid>
        </form>
      </div>
    );
  }
}

export default withStyles(styles)(SignUpFormComponent);
