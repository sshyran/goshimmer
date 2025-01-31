import * as React from 'react';
import Container from "react-bootstrap/Container";
import NodeStore from "app/stores/NodeStore";
import { inject, observer } from "mobx-react";
import ExplorerStore from "app/stores/ExplorerStore";
import ListGroup from "react-bootstrap/ListGroup";
import Badge from "react-bootstrap/Badge";
import {resolveBase58BranchID} from "app/utils/branch";


interface Props {
    nodeStore?: NodeStore;
    explorerStore?: ExplorerStore;
    match?: {
        params: {
            id: string,
        }
    }
}

@inject("nodeStore")
@inject("explorerStore")
@observer
export class ExplorerBranchQueryResult extends React.Component<Props, any> {
    componentDidMount() {
        this.props.explorerStore.getBranch(this.props.match.params.id);
        this.props.explorerStore.getBranchChildren(this.props.match.params.id);
        this.props.explorerStore.getBranchConflicts(this.props.match.params.id);
    }

    componentWillUnmount() {
        this.props.explorerStore.reset();
    }
    render() {
        let {id} = this.props.match.params;
        let { query_err, branch, branchChildren, branchConflicts } = this.props.explorerStore;

        if (query_err) {
            return (
                <Container>
                    <h4>Branch not found - 404</h4>
                    <span>{id}</span>
                </Container>
            );
        }
        let renderInclusionState = (inclusionState: string) => {
            let variant = "secondary";
            let value = ""
            switch(inclusionState) {
                case "InclusionState(Confirmed)":
                    variant = "success";
                    value = "confirmed"
                    break;
                case "InclusionState(Rejected)":
                    variant = "danger";
                    value = "rejected"
                    break;
                case "InclusionState(Pending)":
                    variant = "warning";
                    value = "pending"
                    break;
            }
            return <Badge variant={variant}>{value}</Badge>
        }
        return (
            <Container>
                <h4>Branch</h4>
                {branch && <ListGroup>
                    <ListGroup.Item>ID: {resolveBase58BranchID(branch.id)}</ListGroup.Item>
                    <ListGroup.Item>Type: {branch.type}</ListGroup.Item>
                    <ListGroup.Item>Parents:
                        <ListGroup>
                        {branch.parents.map((p,i) => <ListGroup.Item key={i}><a href={`/explorer/branch/${p}`}>{resolveBase58BranchID(p)}</a></ListGroup.Item>)}
                        </ListGroup>
                    </ListGroup.Item>
                    <ListGroup.Item>Conflicts:
                        {branch.conflictIDs && <ListGroup>
                            {branch.conflictIDs.map((c,i) => <ListGroup.Item key={i}><a href={`/explorer/output/${c}`}>{c}</a></ListGroup.Item>)}
                        </ListGroup>}
                    </ListGroup.Item>
                    <ListGroup.Item>Finalized: {branch.finalized.toString()}</ListGroup.Item>
                    <ListGroup.Item>Monotonically Liked: {branch.monotonicallyLiked.toString()}</ListGroup.Item>
                    <ListGroup.Item>Inclusion State: {renderInclusionState(branch.inclusionState)}</ListGroup.Item>
                    <ListGroup.Item> Children:
                        {branchChildren && <ListGroup>
                            {branchChildren.childBranches.map((c,i) => <ListGroup.Item key={i}><a href={`/explorer/branch/${c.branchID}`}>{resolveBase58BranchID(c.branchID)}</a></ListGroup.Item>)}
                        </ListGroup> }
                    </ListGroup.Item>
                    <ListGroup.Item> Conflicts:
                        {branchConflicts && <ListGroup>
                            {branchConflicts.conflicts.map((c,i) => <div key={i}>
                                OutputID: <a href={`/explorer/output/${c.outputID.base58}`}>{c.outputID.base58}</a>
                                <ListGroup className={"mb-2"}>
                                    {c.branchIDs.map((b,j) => <ListGroup.Item key={j}>
                                        <a href={`/explorer/branch/${b}`}>{resolveBase58BranchID(b)}</a>
                                    </ListGroup.Item>)}
                                </ListGroup>
                            </div>)}
                        </ListGroup> }
                    </ListGroup.Item>
                </ListGroup>}
            </Container>
        )
    }
}