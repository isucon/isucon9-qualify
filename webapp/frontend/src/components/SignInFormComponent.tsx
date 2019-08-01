import React from 'react';

import {Avatar, Typography, TextField, Button, Grid, createStyles, WithStyles} from '@material-ui/core';
import { Theme } from '@material-ui/core/styles/createMuiTheme';
import { LockOutlined } from '@material-ui/icons';
import { Link as RouteLink } from 'react-router-dom';
import {StyleRules} from "@material-ui/core/styles";
import withStyles from "@material-ui/core/styles/withStyles";

const styles = (theme: Theme): StyleRules => createStyles({
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
    onSubmit: (userId: string, password: string) => void
}

interface SignInFormComponentState {
    userId: string,
    password: string,
}

class SignInPageFormComponent extends React.Component<SignInFormComponentProps, SignInFormComponentState> {
    constructor(props: SignInFormComponentProps) {
        super(props);

        this.state = {
            userId: '',
            password: '',
        };

        this._onSubmit = this._onSubmit.bind(this);
        this._onChangeUserId = this._onChangeUserId.bind(this);
        this._onChangePassword = this._onChangePassword.bind(this);
    }

    _onSubmit(e: React.MouseEvent) {
        e.preventDefault();
        const { userId, password } = this.state;
        this.props.onSubmit(userId, password);
    }

    _onChangeUserId(e: React.ChangeEvent<HTMLInputElement>) {
        this.setState({
            userId: e.target.value
        })
    }

    _onChangePassword(e: React.ChangeEvent<HTMLInputElement>) {
        this.setState({
            password: e.target.value
        })
    }

    render() {
        const { userId, password } = this.state;
        const { classes } = this.props;

        return (
            <div>
                <Avatar className={classes.avatar}>
                    <LockOutlined/>
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
                        id="id"
                        label="ログインID"
                        name="id"
                        autoFocus
                        value={userId}
                        onChange={this._onChangeUserId}
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
                            <RouteLink to="/signup">新規登録はこちら</RouteLink>
                        </Grid>
                    </Grid>
                </form>
            </div>
        );
    }
}

export default withStyles(styles)(SignInPageFormComponent);
