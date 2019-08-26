import React from 'react';
import BasePageContainer from '../containers/BasePageContainer';
import { routes } from '../routes/Route';
import { Button, Theme } from '@material-ui/core';
import Typography from '@material-ui/core/Typography';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { Link, LinkProps } from 'react-router-dom';

const useStyles = makeStyles((theme: Theme) => ({
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
  checklist: {
    marginTop: theme.spacing(2),
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
    (props: LinkProps, ref: React.Ref<any>) => (
      <Link innerRef={ref} {...props}>
        ログイン
      </Link>
    ),
  );
  const RegisterButtonLink = React.forwardRef(
    (props: LinkProps, ref: React.Ref<any>) => (
      <Link innerRef={ref} {...props}>
        新規会員登録
      </Link>
    ),
  );

  return (
    <BasePageContainer>
      <div className={classes.paper}>
        <img className={classes.img} src={'/logo.png'} alt={'ISUCARI'} />
        <div className={classes.textarea}>
          <Typography variant="h6">
            椅子限定のフリマサイト ついにリリース！
          </Typography>
          <div className={classes.checklist}>
            <Typography variant="h6">✔ 安全なカード決済</Typography>
            <Typography variant="h6">✔ お互い匿名で安心配送</Typography>
          </div>
          <Typography variant="h6">
            安心安全にあなただけの椅子を手に入れよう！
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
