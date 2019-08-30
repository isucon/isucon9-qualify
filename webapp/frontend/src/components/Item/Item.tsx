import React from 'react';
import Card from '@material-ui/core/Card';
import { Link as RouteLink } from 'react-router-dom';
import GridListTileBar from '@material-ui/core/GridListTileBar';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { routes } from '../../routes/Route';
import { Theme } from '@material-ui/core';
import { ItemStatus } from '../../dataObjects/item';
import { ItemImage } from '../ItemImage';
import GridListTile from '@material-ui/core/GridListTile';

const useStyles = makeStyles((theme: Theme) => ({
  card: {
    width: '300px',
    position: 'relative',
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
    <RouteLink to={routes.item.getPath(itemId)}>
      <Card className={classes.card}>
        <GridListTile>
          <ItemImage
            imageUrl={imageUrl}
            title={title}
            isSoldOut={status !== 'on_sale'}
            width={300}
          />
          <GridListTileBar title={title} subtitle={`${price}ｲｽｺｲﾝ`} />
        </GridListTile>
      </Card>
    </RouteLink>
  );
};

export { Item };
