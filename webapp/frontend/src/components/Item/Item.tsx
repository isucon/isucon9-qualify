import React from 'react';
import Card from '@material-ui/core/Card';
import { Link as RouteLink } from 'react-router-dom';
import GridListTileBar from '@material-ui/core/GridListTileBar';
import makeStyles from '@material-ui/core/styles/makeStyles';
import { routes } from '../../routes/Route';
import {Theme} from "@material-ui/core";
import {ItemStatus} from "../../dataObjects/item";

const useStyles = makeStyles((theme: Theme) => ({
  itemImage: {
    height: '100%',
  },
}));

interface Props {
  itemId: number;
  imageUrl: string;
  title: string;
  price: number;
  status: ItemStatus,
}

const Item: React.FC<Props> = ({
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

export { Item };
