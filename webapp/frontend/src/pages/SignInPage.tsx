import React from 'react';

import { Avatar, Typography, TextField, Button, Grid } from '@material-ui/core';
import { LockOutlined } from '@material-ui/icons';
import { Link as RouteLink } from 'react-router-dom';



const SignInPage: React.FC = () => (
    <div>
        <Avatar>
            <LockOutlined/>
        </Avatar>
        <Typography component="h1" variant="h5">
            ログインページ
        </Typography>
        <form noValidate>
            <TextField
                variant="outlined"
                margin="normal"
                required
                fullWidth
                id="id"
                label="ログインID"
                name="id"
                autoFocus
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
            />
            <Button
                type="submit"
                fullWidth
                variant="contained"
                color="primary"
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

export { SignInPage }