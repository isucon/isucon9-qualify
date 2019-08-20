import * as React from 'react';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { AppBar } from '@material-ui/core';
import Drawer from '@material-ui/core/Drawer';
import List from '@material-ui/core/List';
import ListItem from '@material-ui/core/ListItem';
import ListItemText from '@material-ui/core/ListItemText';
import Toolbar from '@material-ui/core/Toolbar';
import IconButton from '@material-ui/core/IconButton';
import MenuIcon from '@material-ui/icons/Menu';
import Typography from '@material-ui/core/Typography';

const useStyles = makeStyles(theme => ({
  //
}));

interface Props {
  isLoggedIn: boolean;
  ownUserId: number;
  goToTopPage: () => void;
  goToUserPage: (userId: number) => void;
  goToSettingPage: () => void;
}

const Header: React.FC<Props> = ({
  isLoggedIn,
  ownUserId,
  goToTopPage,
  goToUserPage,
  goToSettingPage,
}) => {
  const classes = useStyles();
  const [state, setState] = React.useState({
    open: false,
  });

  const onClickTop = (e: React.MouseEvent) => {
    e.preventDefault();
    goToTopPage();
  };

  const onClickMyPage = (e: React.MouseEvent) => {
    e.preventDefault();
    goToUserPage(ownUserId);
  };

  const onClickMySettingPage = (e: React.MouseEvent) => {
    e.preventDefault();
    goToSettingPage();
  };

  const toggleDrawer = (open: boolean) => (event: React.MouseEvent) => {
    event.preventDefault();
    setState({ ...state, open });
  };

  return (
    <React.Fragment>
      {isLoggedIn && (
        <Drawer open={state.open} onClose={toggleDrawer(false)}>
          <List>
            <ListItem button onClick={onClickTop}>
              <ListItemText primary="新着商品" />
            </ListItem>
            <ListItem button onClick={onClickMyPage}>
              <ListItemText primary="マイページ" />
            </ListItem>
            <ListItem button onClick={onClickMySettingPage}>
              <ListItemText primary="設定" />
            </ListItem>
          </List>
        </Drawer>
      )}
      <AppBar position="fixed">
        <Toolbar>
          {isLoggedIn && (
            <IconButton
              color="inherit"
              onClick={toggleDrawer(true)}
              edge="start"
            >
              <MenuIcon />
            </IconButton>
          )}
          <Typography variant="h6" noWrap>
            ヘッダー
          </Typography>
        </Toolbar>
      </AppBar>
    </React.Fragment>
  );
};

export { Header };
