import React from 'react';
import {ItemData} from "../dataObjects/item";
import makeStyles from "@material-ui/core/styles/makeStyles";
import { ItemListComponent } from '../components/ItemListComponent';
import SellingButtonContainer from "../containers/SellingButtonContainer";
import {ErrorProps, PageComponentWithError} from "../hoc/withBaseComponent";
import {BasePageComponent} from "../components/BasePageComponent";
import LoadingComponent from "../components/LoadingComponent";

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
    loading: boolean,
} & ErrorProps

const ItemListPage: React.FC<ItemListPageProps> = ({ items, loading }: ItemListPageProps) => {
    const classes = useStyles();

    return (
        <BasePageComponent>
            {
                loading ? (
                    <LoadingComponent/>
                ) : (
                    <div className={classes.root}>
                        <ItemListComponent items={items}/>
                        <SellingButtonContainer/>
                    </div>
                )
            }
        </BasePageComponent>
    );
};

export default PageComponentWithError<ItemListPageProps>()(ItemListPage);