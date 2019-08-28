import React from 'react';
import Card from '@material-ui/core/Card';
import { Link as RouteLink } from 'react-router-dom';
import GridListTileBar from '@material-ui/core/GridListTileBar';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { routes } from '../../routes/Route';
import { Theme } from '@material-ui/core';
import { ItemStatus } from '../../dataObjects/item';
import GridListTile from '@material-ui/core/GridListTile';
import Typography from '@material-ui/core/Typography';

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    width: '300px',
    position: 'relative',
  },
  itemImage: {
    width: '300px',
    height: 'auto',
  },
  soldOut: {
    position: 'absolute',
    top: 0,
    right: 0,
    zIndex: 1,
    width: 0,
    height: 0,
    borderStyle: 'solid',
    borderWidth: '0 140px 140px 0',
    borderColor: 'transparent #ff0000 transparent transparent;',
  },
  soldOutText: {
    position: 'absolute',
    top: '35px',
    right: '1px',
    fontWeight: theme.typography.fontWeightBold,
    zIndex: 2,
    transform: 'rotate(45deg);',
    color: theme.palette.primary.contrastText,
  },
}));

interface Props {
  itemId: number;
  imageUrl: string;
  title: string;
  price: number;
  status: ItemStatus;
}

const Item: React.FC<Props> = ({ itemId, imageUrl, title, price, status }) => {
  const classes = useStyles();

  return (
    <Card className={classes.card}>
      {status === 'sold_out' && (
        <React.Fragment>
          <div className={classes.soldOut} />
          <Typography className={classes.soldOutText} variant="h6">
            SOLD OUT
          </Typography>
        </React.Fragment>
      )}
      <GridListTile>
        <RouteLink to={routes.item.getPath(itemId)}>
          <img className={classes.itemImage} src={imageUrl} alt={title} />
        </RouteLink>
        <GridListTileBar title={title} subtitle={`${price}ｲｽｺｲﾝ`} />
      </GridListTile>
    </Card>
  );
};

export { Item };
