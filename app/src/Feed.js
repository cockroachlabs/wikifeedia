import React from 'react';
import gql from 'graphql-tag';


import { useQuery } from '@apollo/react-hooks';

const GET_FEED = gql`
query Feed(
  $project: String!, $offset: Int, $limit: Int, $followerRead: Boolean, $asOf: String
) {
  articles(
    project: $project, offset: $offset, limit: $limit, 
    followerRead: $followerRead, asOf: $asOf
  ) {
    asOf
    articles {
      project
      abstract
      article
      articleURL
      dailyViews
      imageURL
      thumbnailURL
      title
    }
  }
}`;

function useFollowerRead() {
    const urlParams = new URLSearchParams(window.location.search);
    return !(urlParams.get("use_follower_read") === "false");
}

export function Feed({ project }) {
  const { loading, err, data, fetchMore } = useQuery(GET_FEED, {
    variables: {
      project: project,
      offset: 0,
      limit: 10,
      followerRead: useFollowerRead()
    },
    fetchPolicy: "cache-and-network"
  });
  const articles = (data !== undefined) ? data : {articles: {articles: []}}
  return (
    <FeedContainer
      key="feed"
      articles={articles}
      loading={loading}
      error={err}
      onLoadMore={() =>
        fetchMore({
          variables: {
            offset: articles.articles.articles.length,
            asOf: articles.articles.asOf,
          },
          updateQuery: (prev, { fetchMoreResult }) => {
            if (!fetchMoreResult) return prev;
            const prevArticles = prev.articles.articles;
            const newArticles = fetchMoreResult.articles.articles;
            const newAsOf = fetchMoreResult.articles.asOf
            return {
              articles: {
                  asOf: newAsOf,
                  articles: [...prevArticles, ...newArticles],
                  __typename: prev.articles.__typename,
              },
              __typename: prev.__typename,
            }
          }
        })
      }
    />
  )
}

class FeedContainer extends React.Component {

  render() {
    if (this.props.error) return `Error! ${this.props.error.message}`;
    if (!this.props.articles && this.props.loading) return <p>Loading....</p>;
    const articles = this.props.articles.articles.articles;
    return (
      <div className="Articles">
        {articles.map(({
          project,
          article,
          abstract,
          articleURL,
          dailyViews,
          imageURL,
          thumbnailURL,
          title
        }, idx) => (
        <div className="Article" id={article} key={idx} project={project}>
          <div className="ArticleImageContainer">
            <div className="ArticleImage">
              <a href={articleURL} target="_blank" rel="noopener noreferrer">
                <img src={thumbnailURL} alt={title}/>
              </a>
            </div>
          </div>
          <div className="ArticleContent">
            <h2 className="ArticleTitle">{title}</h2>
            <p className="ArticleAbstractText">
              {abstract}
            </p>
          </div>
        </div>)
      )}
      </div>
    );
  }

  componentDidMount() {
    window.addEventListener("scroll", this.handleOnScroll);
  }

  componentWillUnmount() {
    window.removeEventListener("scroll", this.handleOnScroll);
  }

  handleOnScroll = () => {
    // http://stackoverflow.com/questions/9439725/javascript-how-to-detect-if-browser-window-is-scrolled-to-bottom
    var scrollTop =
      (document.documentElement && document.documentElement.scrollTop) ||
      document.body.scrollTop;
    var scrollHeight =
      (document.documentElement && document.documentElement.scrollHeight) ||
      document.body.scrollHeight;
    var clientHeight =
      document.documentElement.clientHeight || window.innerHeight;
    var scrolledToBottom = Math.ceil(scrollTop + clientHeight) >= scrollHeight;
    if (scrolledToBottom) {
      this.props.onLoadMore();
    }
  };
}
