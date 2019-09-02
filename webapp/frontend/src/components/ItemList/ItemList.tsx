import React from 'react';
import { TimelineItem } from '../../dataObjects/item';
import makeStyles from '@material-ui/core/styles/makeStyles';
import GridList from '@material-ui/core/GridList';
import { Item } from '../Item';
import GridListTile from '@material-ui/core/GridListTile';
import InfiniteScroll from 'react-infinite-scroller';
import { Theme } from '@material-ui/core';
import { TimelineLoading } from '../TimelineLoading';

const useStyles = makeStyles((theme: Theme) => ({
  gridList: {
    display: 'flex',
    flexWrap: 'wrap',
    justifyContent: 'flex-start',
  },
  grid: {
    height: '100%',
  },
}));

export interface Props {
  items: TimelineItem[];
  hasNext: boolean;
  loadMore: (page: number) => void;
}

const ItemList: React.FC<Props> = function({
  items,
  hasNext,
  loadMore,
}: Props) {
  const classes = useStyles();

  const itemComponents = [];

  for (const item of items) {
    itemComponents.push(
      <GridListTile className={classes.grid} key={item.id}>
        <Item
          itemId={item.id}
          imageUrl={item.thumbnailUrl}
          title={item.name}
          price={item.price}
          status={item.status}
        />
      </GridListTile>,
    );
  }

  return (
    <InfiniteScroll
      pageStart={0}
      loadMore={loadMore}
      hasMore={hasNext}
      loader={<TimelineLoading />}
    >
      <GridList className={classes.gridList} cellHeight="auto" cols={3}>
        {itemComponents}
      </GridList>
    </InfiniteScroll>
  );
};

export { ItemList };
