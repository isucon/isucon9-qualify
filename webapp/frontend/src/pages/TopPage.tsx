import React from 'react';
import BasePageContainer from "../containers/BasePageContainer";
import {routes} from "../routes/Route";
import {Button} from "@material-ui/core";
import Typography from "@material-ui/core/Typography";
import makeStyles from "@material-ui/core/styles/makeStyles";
import {Link, LinkProps} from "react-router-dom";

const useStyles = makeStyles(theme => ({
    paper: {
        marginTop: theme.spacing(2),
        display: 'flex',
        flexDirection: 'column',
        alignItems: 'center',
    },
    textarea: {
        marginTop: theme.spacing(1),
        marginBottom: theme.spacing(2),
    },
    img: {
        width: '70%',
    },
    button: {
        margin: theme.spacing(1),
    },
}));

const TopPage: React.FC = () => {
    const classes = useStyles();
    const LoginButtonLink = React.forwardRef(
        (props: LinkProps, ref: React.Ref<any>) => <Link innerRef={ref} {...props}>ログイン</Link>
    );
    const RegisterButtonLink = React.forwardRef(
        (props: LinkProps, ref: React.Ref<any>) => <Link innerRef={ref} {...props}>新規会員登録</Link>
    );

    return (
        <BasePageContainer>
            <div className={classes.paper}>
                <img className={classes.img} src={'/logo.png'} alt={'ISUCARI'}/>
                <div className={classes.textarea}>
                    <Typography variant="h6">
                        ついにリリース！
                    </Typography>
                    <Typography variant="h6">
                        椅子限定C2CのECサービス カードで簡単決済。
                    </Typography>
                    <Typography variant="h6">
                        もちろんセキュリティも万全。 お互いの住所を知らなくても配送可能。
                    </Typography>
                </div>
                <Button
                    color="primary"
                    fullWidth
                    className={classes.button}
                    variant="contained"
                    size="medium"
                    component={LoginButtonLink}
                    to={routes.login.path}
                />
                <Button
                    color="primary"
                    fullWidth
                    className={classes.button}
                    variant="outlined"
                    size="medium"
                    component={RegisterButtonLink}
                    to={routes.register.path}
                />
            </div>
        </BasePageContainer>
    );
};

export default TopPage;