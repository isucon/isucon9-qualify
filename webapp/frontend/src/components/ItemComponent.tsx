import React from 'react';
import Card from '@material-ui/core/Card';
import { Link as RouteLink } from 'react-router-dom';
import GridListTileBar from '@material-ui/core/GridListTileBar';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { routes } from '../routes/Route';

const useStyles = makeStyles(theme => ({
  itemImage: {
    height: '100%',
  },
}));

interface ItemComponentProps {
  itemId: number;
  imageUrl: string;
  title: string;
  price: number;
}

const ItemComponent: React.FC<ItemComponentProps> = ({
  itemId,
  imageUrl,
  title,
  price,
}) => {
  const classes = useStyles();

  return (
    <Card>
      <RouteLink to={routes.item.getPath(itemId)}>
        <img className={classes.itemImage} src={imageUrl} alt={title} />
      </RouteLink>
      <GridListTileBar title={title} subtitle={`${price}ｲｽｺｲﾝ`} />
    </Card>
  );
};

export { ItemComponent };
