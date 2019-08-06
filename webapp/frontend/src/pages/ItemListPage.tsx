import React from 'react';
import {ItemData} from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import { ItemListComponent } from '../components/ItemListComponent';
import SellingButtonContainer from "../containers/SellingButtonContainer";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";
import {mockItems} from "../mocks";
import {BasePageComponent} from "../components/BasePageComponent";

const useStyles = makeStyles(theme => ({
    root: {
        display: 'flex',
        flexWrap: 'wrap',
        marginTop: theme.spacing(1),
        justifyContent: 'space-around',
        overflow: 'hidden',
    },
}));

type ItemListPageProps = {
    items: ItemData[],
} & ErrorProps

const ItemListPage: React.FC/*<ItemListPageProps>*/ = (/*{ items }: ItemListPageProps*/) => {
    const classes = useStyles();
    const items = mockItems;

    return (
        <BasePageComponent>
            <div className={classes.root}>
                <ItemListComponent items={items}/>
                <SellingButtonContainer />
            </div>
        </BasePageComponent>
    );
};

export default PageComponentWithError<ItemListPageProps>()(ItemListPage);